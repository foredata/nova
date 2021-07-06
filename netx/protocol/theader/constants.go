package theader

type IntMap = map[int]string

const (
	commonHeaderSize = 10         // MAGIC +FLAGS+ + SEQUENCE NUMBER+ Header Size
	headerMagic      = 0x0FFF0000 //
	magicMask        = 0xFFFF0000
	// flagsMask        = 0x0000FFFF
	maxFrameSize  = 0x3FFFFFFF
	maxHeaderSize = 131071
)

const (
	protoIDBinary  = 0
	protoIDCompact = 2
)

// TransformID Numerical ID of transform function
type TransformID uint32

const (
	// TransformNone Default null transform
	TransformNone TransformID = 0
	// TransformZlib Apply zlib compression
	TransformZlib TransformID = 1
	// TransformHMAC Deprecated and no longer supported
	TransformHMAC TransformID = 2
	// TransformSnappy Apply snappy compression
	TransformSnappy TransformID = 3
	// TransformQLZ Deprecated and no longer supported
	TransformQLZ TransformID = 4
	// TransformZstd Apply zstd compression
	TransformZstd TransformID = 5
)

type infoIDType uint32

const (
	infoIDPadding     infoIDType = 0x00
	infoIDKeyValue    infoIDType = 0x01
	infoIDIntKeyValue infoIDType = 0x10
)

// const (
// 	intKeyTransportType  = 1
// 	intKeyLogID          = 2
// 	intKeyFromService    = 3
// 	intKeyFromCluster    = 4
// 	intKeyFromIDC        = 5
// 	intKeyToService      = 6
// 	intKeyToCluster      = 7
// 	intKeyToIDC          = 8
// 	intKeyToMethod       = 9
// 	intKeyEnv            = 10
// 	intKeyDestAddr       = 11
// 	intKeyRpcTimeout     = 12
// 	intKeyRingHashKey    = 14
// 	intKeyWithHeader     = 16
// 	intKeyConnTimeout    = 17
// 	intKeyTraceSpanCtx   = 18
// 	intKeyShortConn      = 19
// 	intKeyFromMethod     = 20
// 	intKeyStressTag      = 21
// 	intKeyMsgType        = 22
// 	intKeyConnRecycle    = 23
// 	intKeyRawRingHashKey = 24
// 	intKeyLbType         = 25
// 	intKeyClusterShardId = 26
// )
