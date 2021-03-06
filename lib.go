package BasicHashMapNoPointer

var INITIAL_CAPACITY = 1 << 1
var INITIAL_BUCKET_SIZE = 1 << 1
var THRESHOLD_MAX float32 = 6.0
var THRESHOLD_MIN float32 = 0.2
var GROW_SHIFT = 1
var EMPTY_VALUE = ^uint32(0)
var BUCKET_END_VALUE = ^uint32(0) - 1

type BasicHashMapNoPointer struct {
	store []uint32
	maxSize int
	buckets int
	Size int
}

func createByCapacity(capacity int) BasicHashMapNoPointer {
	hm := BasicHashMapNoPointer{
		make([]uint32, 2 * capacity * INITIAL_BUCKET_SIZE + capacity * 2),
		capacity * INITIAL_BUCKET_SIZE,
		capacity,
		0,
	}
	for i := hm.buckets; i < len(hm.store); i++ {
		hm.store[i] = EMPTY_VALUE
	}
	for i := 0; i < hm.buckets; i++ {
		hm.store[i] = uint32(hm.buckets + i * (2 * INITIAL_BUCKET_SIZE + 1))
		hm.store[hm.buckets + (i + 1) * (2 * INITIAL_BUCKET_SIZE + 1) - 1] = BUCKET_END_VALUE
	}
	return hm
}

func (hm BasicHashMapNoPointer) hashFunc(key uint32) uint32 {
	return (key ^ (key >> 20) ^ (key >> 12) ^ (key >> 7) ^ (key >> 4)) & uint32(hm.buckets - 1)
}

func New() BasicHashMapNoPointer {
	return createByCapacity(INITIAL_CAPACITY)
}

func (hm *BasicHashMapNoPointer) Put(key uint32, value uint32) {
	hm.put(key, value, true)
}

func (hm *BasicHashMapNoPointer) put(key uint32, value uint32, resize bool) {
	bucket := hm.hashFunc(key)
	bucketStart := hm.store[bucket]

	for i := bucketStart; i < uint32(len(hm.store)); i += 2 {
		if hm.store[i] == key {
			hm.store[i + 1] = value
			return
		} else if hm.store[i] == EMPTY_VALUE {
			hm.store[i] = key
			hm.store[i + 1] = value
			hm.changeSizeBy(1, resize)
			return
		} else if hm.store[i] == BUCKET_END_VALUE {
			// bucket full, resize bucket
			newStore := make([]uint32, len(hm.store) + (INITIAL_BUCKET_SIZE * 2))
			copy(newStore[0:i], hm.store[0:i])
			for k := i; k < i + uint32(INITIAL_BUCKET_SIZE * 2); k += 2 {
				newStore[k] = EMPTY_VALUE
			}
			newStore[i + uint32(INITIAL_BUCKET_SIZE * 2)] = BUCKET_END_VALUE
			for k := int(bucket) + 1; k < hm.buckets; k++ {
				newStore[k] += uint32(INITIAL_BUCKET_SIZE * 2)
			}
			for k := int(bucket) + 1; k < hm.buckets - 1; k++ {
				copy(newStore[newStore[k]:newStore[k+1]], hm.store[hm.store[k]:hm.store[k+1]])
			}
			copy(newStore[newStore[hm.buckets - 1]:], hm.store[hm.store[hm.buckets - 1]:])
			newStore[i] = key
			newStore[i + 1] = value
			hm.store = newStore
			hm.maxSize += INITIAL_BUCKET_SIZE
			hm.changeSizeBy(1, resize)
			return
		}
	}
}

func (hm *BasicHashMapNoPointer) Get(key uint32) uint32 {
	bucketStart := hm.store[hm.hashFunc(key)]

	for i := bucketStart; hm.store[i] != BUCKET_END_VALUE && hm.store[i] != EMPTY_VALUE; i += 2 {
		if hm.store[i] == key {
			return hm.store[i + 1]
		}
	}
	return 0
}

func (hm *BasicHashMapNoPointer) GetIndex(key uint32) uint32 {
	bucketStart := hm.store[hm.hashFunc(key)]

	for i := bucketStart; hm.store[i] != BUCKET_END_VALUE && hm.store[i] != EMPTY_VALUE; i += 2 {
		if hm.store[i] == key {
			return i
		}
	}
	return 0
}

func (hm *BasicHashMapNoPointer) PutIndex(keyIndex uint32, value uint32) {
	hm.store[keyIndex + 1] = value
}

func max(a, b uint32) uint32 {
	if a < b {
		return b
	}
	return a
}

func (hm *BasicHashMapNoPointer) UpdateMax(key uint32, value uint32) {
	storeIndex := hm.GetIndex(key)
	tmp := value
	if int(storeIndex) >= hm.buckets {
		tmp = max(hm.store[storeIndex + 1], value)
		hm.PutIndex(storeIndex, tmp)
	} else {
		hm.Put(key, value)
	}
}

func (hm *BasicHashMapNoPointer) Add(key uint32, value uint32) {
	storeIndex := hm.GetIndex(key)
	if int(storeIndex) >= hm.buckets {
		hm.PutIndex(storeIndex, hm.store[storeIndex + 1] + value)
	} else {
		hm.Put(key, value)
	}
}

func (hm *BasicHashMapNoPointer) Store() ([]uint32, int) {
	return hm.store, hm.buckets
}

func (hm *BasicHashMapNoPointer) Iter() []uint32 {
	pairs := make([]uint32, hm.Size * 2)
	index := 0

	for i := 0; i < hm.buckets; i++ {
		for k := hm.store[i]; k < uint32(len(hm.store)); k += 2 {
			if hm.store[k] == EMPTY_VALUE || hm.store[k] == BUCKET_END_VALUE {
				copy(pairs[index:], hm.store[hm.store[i]:k])
				index += int(k - hm.store[i])
				break
			}
		}
	}

	return pairs
}

func (hm *BasicHashMapNoPointer) Len() int {
	return hm.Size
}

func (hm *BasicHashMapNoPointer) Clone() BasicHashMapNoPointer {
	newStore := make([]uint32, len(hm.store))
	copy(newStore, hm.store)
	return BasicHashMapNoPointer{
		newStore,
		hm.maxSize,
		hm.buckets,
		hm.Size,
	}
}

func (hm *BasicHashMapNoPointer) changeSizeBy(change int, resize bool) {
	hm.Size += change
	if resize {
		hm.resizeOnThreshold()
	}
}

func (hm *BasicHashMapNoPointer) resizeOnThreshold() {
	newSize := 0
	if hm.Size < int(float32(hm.buckets) * THRESHOLD_MIN) && hm.Size > INITIAL_CAPACITY {
		newSize = hm.buckets >> 1
	} else if hm.Size > int(float32(hm.buckets) * THRESHOLD_MAX) {
		newSize = hm.buckets << GROW_SHIFT
	} else {
		return
	}

	newHashMap := createByCapacity(newSize)
	iter := hm.Iter()
	for i := 0; i < len(iter); i += 2 {
		newHashMap.put(iter[i], iter[i + 1], false)
	}

	*hm = newHashMap
}

