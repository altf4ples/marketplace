package cart

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const cartTTL = 7 * 24 * time.Hour

type CartRepository interface {
	GetOrCreate(context.Context, uuid.UUID) (Cart, error)
	Get(context.Context, uuid.UUID) (Cart, error)
	AddItem(context.Context, uuid.UUID, uuid.UUID, int64) (Cart, error)
	RemoveItem(context.Context, uuid.UUID, uuid.UUID) (Cart, error)
	Clear(context.Context, uuid.UUID) error
	Take(context.Context, uuid.UUID) ([]CartItem, error)
	Restore(context.Context, uuid.UUID, []CartItem) error
}

type CartRepo struct{ rdb *redis.Client }

func NewCartRepo(rdb *redis.Client) *CartRepo { return &CartRepo{rdb: rdb} }
func cartKey(id uuid.UUID) string             { return "cart:" + id.String() }

func (r *CartRepo) read(ctx context.Context, userID uuid.UUID) (Cart, error) {
	values, err := r.rdb.HGetAll(ctx, cartKey(userID)).Result()
	if err != nil {
		return Cart{}, err
	}
	items := make([]CartItem, 0, len(values))
	for rawID, rawQty := range values {
		id, err := uuid.Parse(rawID)
		if err != nil {
			return Cart{}, err
		}
		qty, err := strconv.ParseInt(rawQty, 10, 64)
		if err != nil {
			return Cart{}, err
		}
		items = append(items, CartItem{ProductID: id, Quantity: qty})
	}
	return Cart{UserID: userID, Items: items}, nil
}

func (r *CartRepo) GetOrCreate(ctx context.Context, userID uuid.UUID) (Cart, error) {
	c, err := r.read(ctx, userID)
	if err != nil {
		return Cart{}, err
	}
	if len(c.Items) > 0 {
		if err := r.rdb.Expire(ctx, cartKey(userID), cartTTL).Err(); err != nil {
			return Cart{}, err
		}
	}
	return c, nil
}
func (r *CartRepo) Get(ctx context.Context, userID uuid.UUID) (Cart, error) {
	return r.read(ctx, userID)
}

var addItemScript = redis.NewScript(`
local current = tonumber(redis.call('HGET', KEYS[1], ARGV[1]) or '0')
local next = current + tonumber(ARGV[2])
if next > 1000 then return -1 end
redis.call('HINCRBY', KEYS[1], ARGV[1], ARGV[2])
redis.call('EXPIRE', KEYS[1], ARGV[3])
return next
`)

func (r *CartRepo) AddItem(ctx context.Context, userID, productID uuid.UUID, qty int64) (Cart, error) {
	key := cartKey(userID)
	next, err := addItemScript.Run(ctx, r.rdb, []string{key}, productID.String(), qty, int64(cartTTL.Seconds())).Int64()
	if err != nil {
		return Cart{}, err
	}
	if next < 0 {
		return Cart{}, ErrQuantityInvalid
	}
	return r.read(ctx, userID)
}
func (r *CartRepo) RemoveItem(ctx context.Context, userID, productID uuid.UUID) (Cart, error) {
	key := cartKey(userID)
	if err := r.rdb.HDel(ctx, key, productID.String()).Err(); err != nil {
		return Cart{}, err
	}
	return r.read(ctx, userID)
}
func (r *CartRepo) Clear(ctx context.Context, userID uuid.UUID) error {
	return r.rdb.Del(ctx, cartKey(userID)).Err()
}

// Take atomically reads and deletes the cart. A concurrent HINCRBY is ordered
// either before the script (included in this order) or after it (left in a new cart).
var takeCartScript = redis.NewScript(`
local values = redis.call('HGETALL', KEYS[1])
redis.call('DEL', KEYS[1])
return values
`)

func (r *CartRepo) Take(ctx context.Context, userID uuid.UUID) ([]CartItem, error) {
	raw, err := takeCartScript.Run(ctx, r.rdb, []string{cartKey(userID)}).Slice()
	if err != nil {
		return nil, err
	}
	items := make([]CartItem, 0, len(raw)/2)
	for i := 0; i+1 < len(raw); i += 2 {
		rawID, ok := raw[i].(string)
		if !ok {
			return nil, fmt.Errorf("unexpected Redis cart product id type %T", raw[i])
		}
		id, err := uuid.Parse(rawID)
		if err != nil {
			return nil, err
		}
		var qty int64
		switch v := raw[i+1].(type) {
		case string:
			qty, err = strconv.ParseInt(v, 10, 64)
		case int64:
			qty = v
		case []byte:
			qty, err = strconv.ParseInt(string(v), 10, 64)
		default:
			err = strconv.ErrSyntax
		}
		if err != nil {
			return nil, err
		}
		items = append(items, CartItem{ProductID: id, Quantity: qty})
	}
	return items, nil
}

// Restore merges a previously claimed cart into the current cart. It deliberately
// increments fields instead of replacing the hash, preserving concurrent additions.
var restoreCartScript = redis.NewScript(`
for i=1,#ARGV-1,2 do
  redis.call('HINCRBY', KEYS[1], ARGV[i], ARGV[i+1])
end
if redis.call('HLEN', KEYS[1]) > 0 then redis.call('EXPIRE', KEYS[1], ARGV[#ARGV]) end
return 1
`)

func (r *CartRepo) Restore(ctx context.Context, userID uuid.UUID, items []CartItem) error {
	if len(items) == 0 {
		return nil
	}
	args := make([]interface{}, 0, len(items)*2+1)
	for _, it := range items {
		args = append(args, it.ProductID.String(), it.Quantity)
	}
	args = append(args, int64(cartTTL.Seconds()))
	return restoreCartScript.Run(ctx, r.rdb, []string{cartKey(userID)}, args...).Err()
}
