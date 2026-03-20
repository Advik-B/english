package types

// ArrayValue is a homogeneous array: every element must share the same TypeKind.
// The zero value has ElementType == TypeUnknown.
type ArrayValue struct {
	ElementType TypeKind
	Elements    []interface{}
}

// LookupTableValue is an ordered dictionary that maps hashable keys (number,
// text, boolean) to values of any type.
type LookupTableValue struct {
	Entries  map[string]interface{} // serialised key (SerializeKey) → value
	KeyOrder []string               // insertion-order serialised keys
}

// NewLookupTable returns an initialised, empty LookupTableValue.
func NewLookupTable() *LookupTableValue {
	return &LookupTableValue{
		Entries:  make(map[string]interface{}),
		KeyOrder: []string{},
	}
}

// Set inserts or updates an entry.  key must already be serialised via SerializeKey.
func (lt *LookupTableValue) Set(serialKey string, value interface{}) {
	if _, exists := lt.Entries[serialKey]; !exists {
		lt.KeyOrder = append(lt.KeyOrder, serialKey)
	}
	lt.Entries[serialKey] = value
}

// Delete removes an entry.  Returns true if the key was present.
func (lt *LookupTableValue) Delete(serialKey string) bool {
	if _, exists := lt.Entries[serialKey]; !exists {
		return false
	}
	delete(lt.Entries, serialKey)
	newOrder := make([]string, 0, len(lt.KeyOrder))
	for _, k := range lt.KeyOrder {
		if k != serialKey {
			newOrder = append(newOrder, k)
		}
	}
	lt.KeyOrder = newOrder
	return true
}

// RangeValue represents an immutable range with lazy evaluation.
// For ranges with more than 20 elements, values are generated on-demand.
type RangeValue struct {
	Start     float64
	End       float64
	Step      float64 // custom step value, default is 1 or -1
	Ascending bool
	cache     []interface{} // cached elements (up to 20 at a time)
	cachePos  int           // position in the range where cache starts
}

// NewRange creates a new RangeValue with default step (1 or -1 based on direction).
func NewRange(start, end float64) *RangeValue {
	step := 1.0
	ascending := start <= end
	if !ascending {
		step = -1.0
	}
	return &RangeValue{
		Start:     start,
		End:       end,
		Step:      step,
		Ascending: ascending,
		cache:     nil,
		cachePos:  0,
	}
}

// NewRangeWithStep creates a new RangeValue with a custom step value.
func NewRangeWithStep(start, end, step float64) *RangeValue {
	// Determine direction based on step sign
	ascending := step > 0
	return &RangeValue{
		Start:     start,
		End:       end,
		Step:      step,
		Ascending: ascending,
		cache:     nil,
		cachePos:  0,
	}
}

// Length returns the total number of elements in the range.
func (r *RangeValue) Length() int {
	if r.Step == 0 {
		return 0 // avoid infinite loop
	}

	if r.Ascending {
		if r.End < r.Start {
			return 0
		}
		// For ascending ranges: count = floor((end - start) / step) + 1
		count := int((r.End-r.Start)/r.Step) + 1
		if count < 0 {
			return 0
		}
		return count
	} else {
		if r.End > r.Start {
			return 0
		}
		// For descending ranges: count = floor((start - end) / abs(step)) + 1
		count := int((r.Start-r.End)/(-r.Step)) + 1
		if count < 0 {
			return 0
		}
		return count
	}
}

// Get returns the element at the given index (0-based).
// Implements lazy evaluation by caching 20 elements at a time.
func (r *RangeValue) Get(index int) (interface{}, bool) {
	length := r.Length()
	if index < 0 || index >= length {
		return nil, false
	}

	// For small ranges (<=20 elements), generate all at once
	if length <= 20 {
		if r.cache == nil {
			r.cache = r.generateChunk(0, length)
			r.cachePos = 0
		}
		return r.cache[index], true
	}

	// For large ranges, use chunked lazy evaluation
	chunkStart := (index / 20) * 20
	if r.cache == nil || r.cachePos != chunkStart {
		chunkSize := 20
		if chunkStart+chunkSize > length {
			chunkSize = length - chunkStart
		}
		r.cache = r.generateChunk(chunkStart, chunkSize)
		r.cachePos = chunkStart
	}

	return r.cache[index-r.cachePos], true
}

// generateChunk generates a chunk of elements starting at offset with the given size.
func (r *RangeValue) generateChunk(offset, size int) []interface{} {
	result := make([]interface{}, size)
	startVal := r.Start + float64(offset)*r.Step
	for i := 0; i < size; i++ {
		result[i] = startVal + float64(i)*r.Step
	}
	return result
}

// ToSlice converts the entire range to a slice (for iteration).
// This is used when the range is small or when full materialization is needed.
func (r *RangeValue) ToSlice() []interface{} {
	length := r.Length()
	result := make([]interface{}, length)
	for i := 0; i < length; i++ {
		result[i] = r.Start + float64(i)*r.Step
	}
	return result
}
