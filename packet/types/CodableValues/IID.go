package CodableValues

import "fmt"

// Object == 0: represents number of objects in Structure
// Else:
//   0 index Defined: represents first instance of object (Structure.Object.1)
//   1 index Defined:
//     First Index == 0: represents number of instances of object
//     0 < First Index < N: represents instance of object at the given index (Structure.Object.First Index)
//     Else: Invalid Index Error
//   2 index Defined:
//     First Index == 0 and Second Index == 0: represents the group of every instance of object
//     0 < First Index <= N and 0 < Second Index <= N:  represents the group of instances of object between the given indexes
//     Else: Invalid Index Error

type IID struct {
	Structure   int
	Object      int
	FirstIndex  *int
	SecondIndex *int
	Length      int
}

func NewIID(structure, object int) *IID {
	return &IID{
		Structure:   structure,
		Object:      object,
		FirstIndex:  nil,
		SecondIndex: nil,
		Length:      2,
	}
}

func NewIIDSingleIndex(structure, object, index int) *IID {
	var FirstIndex *int = new(int)
	*FirstIndex = index
	return &IID{
		Structure:   structure,
		Object:      object,
		FirstIndex:  FirstIndex,
		SecondIndex: nil,
		Length:      3,
	}
}

func NewIIDDoubleIndex(structure, object, firstIndex, secondIndex int) *IID {
	var FirstIndex *int = new(int)
	*FirstIndex = firstIndex
	var SecondIndex *int = new(int)
	*SecondIndex = secondIndex
	return &IID{
		Structure:   structure,
		Object:      object,
		FirstIndex:  FirstIndex,
		SecondIndex: SecondIndex,
		Length:      4,
	}
}

func (iid *IID) Encode() string {
	var encoded string = ""
	encoded += EncodeInt(iid.Structure)
	encoded += EncodeInt(iid.Object)
	if iid.FirstIndex != nil {
		encoded += EncodeInt(*iid.FirstIndex)
		if iid.SecondIndex != nil {
			encoded += EncodeInt(*iid.SecondIndex)
		}
	}
	return encoded
}

func (iid *IID) Decode(data string, length *int) (string, error) {
	iid.Length = *length
	rest := ""
	var err error
	iid.Structure, rest, err = DecodeInt(data)
	if err != nil {
		return "", err
	}
	iid.Object, rest, err = DecodeInt(rest)
	if err != nil {
		return "", err
	}
	if iid.Length > 2 {
		iid.FirstIndex = new(int)
		*iid.FirstIndex, rest, err = DecodeInt(rest)
		if err != nil {
			return "", err
		}
	}
	if iid.Length > 3 {
		iid.SecondIndex = new(int)
		*iid.SecondIndex, rest, err = DecodeInt(rest)
		if err != nil {
			return "", err
		}
	}
	return rest, nil
}

func (iid *IID) Equals(other interface{}) bool {
	ActualValue := other.(*IID)
	structureEqual := iid.Structure == ActualValue.Structure
	objectEqual := iid.Object == ActualValue.Object
	firstIndexEqual := (iid.FirstIndex == nil && ActualValue.FirstIndex == nil) || (iid.FirstIndex != nil && ActualValue.FirstIndex != nil && *iid.FirstIndex == *ActualValue.FirstIndex)
	secondIndexEqual := (iid.SecondIndex == nil && ActualValue.SecondIndex == nil) || (iid.SecondIndex != nil && ActualValue.SecondIndex != nil && *iid.SecondIndex == *ActualValue.SecondIndex)
	return structureEqual && objectEqual && firstIndexEqual && secondIndexEqual
}

func (iid *IID) String() string {
	firstIndexValue := "<nil>"
	if iid.FirstIndex != nil {
		firstIndexValue = fmt.Sprintf("%d", *iid.FirstIndex)
	}
	secondIndexValue := "<nil>"
	if iid.SecondIndex != nil {
		secondIndexValue = fmt.Sprintf("%d", *iid.SecondIndex)
	}
	return fmt.Sprintf("IID{%d.%d.%v.%v}", iid.Structure, iid.Object, firstIndexValue, secondIndexValue)
}

func (iid *IID) Copy() *IID{
	return &IID{
		Structure: iid.Structure,
		Object: iid.Object,
		FirstIndex: iid.FirstIndex,
		SecondIndex: iid.SecondIndex,
		Length: iid.Length,
	}
}

func (iid *IID) GenListOfIIDs(MaxIndex int) ([]*IID, bool) {
	var iids []*IID
	if iid.Length == 4 {
		isSecondBiggerOrEqualThanFirst := *iid.FirstIndex <= *iid.SecondIndex
		isOnlyOneZero := (*iid.FirstIndex > 0 && *iid.SecondIndex == 0) || (*iid.FirstIndex == 0 && *iid.SecondIndex > 0)
		areIndexTooBig := *iid.FirstIndex > MaxIndex || *iid.SecondIndex > MaxIndex
		if !isSecondBiggerOrEqualThanFirst || isOnlyOneZero || areIndexTooBig {
			return nil, false
		}
		startIndex := *iid.FirstIndex
		endIndex := *iid.SecondIndex
		if *iid.FirstIndex == 0 && *iid.SecondIndex == 0 {
			startIndex = 1
			endIndex = MaxIndex
		}
		for i := startIndex; i <= endIndex; i++ {
			iids = append(iids, NewIIDSingleIndex(iid.Structure, iid.Object, i))
		}
	} else {
		iids = append(iids, iid)
	}
	return iids, true
}

// func NewNumberOfObjectsIID(structure int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: 0},
// 		FirstIndex: nil,
// 		SecondIndex: nil,
// 	}
// }

// func NewFirstInstanceIID(structure, object int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: object},
// 		FirstIndex: nil,
// 		SecondIndex: nil,
// 	}
// }

// func NewNumberOfInstancesIID(structure, object int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: object},
// 		FirstIndex: &CodableInt{Value: 0},
// 		SecondIndex: nil,
// 	}
// }

// func NewSingleInstanceIID(structure, object, firstIndex int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: object},
// 		FirstIndex: &CodableInt{Value: firstIndex},
// 		SecondIndex: nil,
// 	}
// }

// func NewEveryInstanceIID(structure, object int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: object},
// 		FirstIndex: &CodableInt{Value: 0},
// 		SecondIndex: &CodableInt{Value: 0},
// 	}
// }

// func NewGroupIID(structure, object, firstIndex, secondIndex int) CodableIID {
// 	return CodableIID{
// 		Structure: CodableInt{Value: structure},
// 		Object: CodableInt{Value: object},
// 		FirstIndex: &CodableInt{Value: firstIndex},
// 		SecondIndex: &CodableInt{Value: secondIndex},
// 	}
// }
