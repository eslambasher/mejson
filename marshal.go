package mejson

import (
	"encoding/base64"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"os"
	"reflect"
	"time"
)

type Mejson map[string]interface{}

func Marshal(in interface{}) (interface{}, error) {
    // short circuit for nil
    if in == nil {
        return nil, nil
    }

    if reflect.TypeOf(in).Kind() == reflect.Slice {
        if v, ok := in.([]byte); ok {
            return marshalBinary(bson.Binary{0x00, v}), nil
        }
        v := reflect.ValueOf(in)
        slice := make([]interface{}, v.Len())
        for i := 0; i < v.Len(); i++ {
            slice[i] = v.Index(i).Interface()
        }
        return marshalSlice(slice)
    } else {
        switch v := in.(type) {
        case primitive.M, map[string]interface{}:
            return marshalMap(toBsonM(v))
		case primitive.ObjectID:
			return marshalObjectId(v), nil
		case bson.D:
            return marshalMap(v.Map())
        case bson.Binary:
            return marshalBinary(v), nil
        case time.Time:
            return marshalTime(v), nil
        case bson.MongoTimestamp:
            return marshalTimestamp(v), nil
        case bson.RegEx:
            return marshalRegex(v), nil
		case primitive.DateTime:
			return marshalDate(v), nil
		case primitive.Undefined:
            return marshalUndefined(), nil
        case string, int, int64, bool, float64, uint8, uint32, int32:
            // Added int32 to the existing types
            return v, nil
        default:
            fmt.Fprintf(os.Stderr, "mejson: unknown !!! type: %T\n", v)
            return v, nil
        }
    }
}

func toBsonM(in interface{}) bson.M {
    switch v := in.(type) {
    case primitive.M:
        // Convert primitive.M to bson.M
        bsonM := bson.M{}
        for key, value := range v {
            bsonM[key] = value
        }
        return bsonM
    case map[string]interface{}:
        // Directly convert map[string]interface{} to bson.M
        return bson.M(v)
    case bson.M:
        // Already bson.M, return as is
		return v
    default:
        // Log unexpected type and return an empty bson.M or handle the error appropriately
        fmt.Printf("toBsonM: unexpected type %T, expected map or bson.M\n", v)
        return bson.M{}
    }
}

func marshalUndefined() map[string]interface{} {
    return map[string]interface{}{
        "$undefined": true,
    }
}
func marshalSlice(in []interface{}) (interface{}, error) {
	result := make([]interface{}, len(in))
	for idx, value := range in {
		mejson, err := Marshal(value)
		if err != nil {
			return nil, err
		}
		result[idx] = mejson
	}
	return result, nil
}

func marshalMap(in bson.M) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	for key, value := range in {
		mejson, err := Marshal(value)
		if err != nil {
			return nil, err
		}
		result[key] = mejson
	}
	return result, nil
}

func marshalObjectId(in primitive.ObjectID) map[string]interface{} {
	return map[string]interface{}{"$oid": in.Hex() }
}

func marshalBinary(in bson.Binary) Mejson {
	return map[string]interface{}{
		"$type":   fmt.Sprintf("%x", in.Kind),
		"$binary": base64.StdEncoding.EncodeToString(in.Data),
	}
}

func marshalTime(in time.Time) map[string]interface{} {
	return map[string]interface{}{
		"$date": int(in.UnixNano() / 1e6),
	}
}

func marshalDate(in primitive.DateTime) map[string]interface{} {
	return map[string]interface{}{
		"$date": int64(in),
	}
}

func marshalTimestamp(in bson.MongoTimestamp) map[string]interface{} {
	//{ "$timestamp": { "t": <t>, "i": <i> } }
	seconds, iteration := int32(in>>32), int32(in)
	return map[string]interface{}{
		"$timestamp": bson.M{"t": seconds, "i": iteration},
	}
}

func marshalRegex(in bson.RegEx) map[string]interface{} {
	return map[string]interface{}{
		"$regex":   in.Pattern,
		"$options": in.Options,
	}
}
