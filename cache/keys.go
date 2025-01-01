package cache

const (
	GetPackageTTL     int32 = 24 * 60 * 60
	GetPackageListTTL int32 = 24 * 60 * 60
)

type CacheKey struct {
	Key string
	TTL int32
}

var CacheKeys = struct {
	GetPackage     CacheKey
	GetPackageList CacheKey
}{
	GetPackage:     CacheKey{Key: "packageDetails/%d:", TTL: GetPackageTTL},
	GetPackageList: CacheKey{Key: "conversationDetails:", TTL: GetPackageListTTL},
}
