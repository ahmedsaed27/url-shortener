package ratelimit

const tokenBucketLua = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local window_ms = tonumber(ARGV[2])

local redis_time = redis.call("TIME")
local now_ms = (tonumber(redis_time[1]) * 1000) + math.floor(tonumber(redis_time[2]) / 1000)

local state = redis.call("HMGET", key, "tokens", "timestamp")
local tokens = tonumber(state[1])
local last_refill_ms = tonumber(state[2])

if tokens == nil then
	tokens = capacity
end
if last_refill_ms == nil then
	last_refill_ms = now_ms
end

local elapsed_ms = math.max(0, now_ms - last_refill_ms)
tokens = math.min(capacity, tokens + (elapsed_ms * capacity / window_ms))

local allowed = 0
local retry_after_ms = 0
if tokens >= 1 then
	allowed = 1
	tokens = tokens - 1
else
	retry_after_ms = math.ceil((1 - tokens) * window_ms / capacity)
end

redis.call("HSET", key, "tokens", tostring(tokens), "timestamp", tostring(now_ms))
redis.call("PEXPIRE", key, math.ceil(window_ms * 2))

return {allowed, math.floor(tokens), retry_after_ms}
`
