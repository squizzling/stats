package kstat

type Type uint

const (
	KstRaw = Type(iota)
	KstNamed
	KstIntr
	KstIo
	KstTimer
)

type Data uint

const (
	KsdChar = Data(iota)
	KsdInt32
	KsdUint32
	KsdInt64
	KsdUint64
	KsdLong
	KsdUlong
	KsdString
)

type Flag int

const (
	KsfVirtual     = Flag(0x01)
	KsfVarSize     = Flag(0x02)
	KsfWritable    = Flag(0x04)
	KsfPersistent  = Flag(0x08)
	KsfDormant     = Flag(0x10)
	KsfInvalid     = Flag(0x20)
	KsfLongStrings = Flag(0x40)
	KsfNoHeaders   = Flag(0x80)
)

type Kstat struct { // struct kstat_s
	Id           int    // kid_t kid
	Type         Type   // uchar_t ks_type
	Flags        Flag   // uchar_t ks_flags
	RecordCount  uint   // uint_t ks_ndata
	DataSize     uint   // size_t ks_data_size
	CreationTime uint64 // hrtime_t ks_crtime
	SnapTime     uint64 // hrtime_t ks_snaptime

	UValues map[string]uint64
	SValues map[string]int64
}
