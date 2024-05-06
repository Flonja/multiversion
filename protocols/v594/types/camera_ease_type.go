package types

import (
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"strings"
)

func CameraEase(s string) uint8 {
	switch strings.ToLower(s) {
	case "linear":
		return protocol.EasingTypeLinear
	case "spring":
		return protocol.EasingTypeSpring
	case "in_quad":
		return protocol.EasingTypeInQuad
	case "out_quad":
		return protocol.EasingTypeOutQuad
	case "in_out_quad":
		return protocol.EasingTypeInOutQuad
	case "in_cubic":
		return protocol.EasingTypeInCubic
	case "out_cubic":
		return protocol.EasingTypeOutCubic
	case "in_out_cubic":
		return protocol.EasingTypeInOutCubic
	case "in_quart":
		return protocol.EasingTypeInQuart
	case "out_quart":
		return protocol.EasingTypeOutQuart
	case "in_out_quart":
		return protocol.EasingTypeInOutQuart
	case "in_quint":
		return protocol.EasingTypeInQuint
	case "out_quint":
		return protocol.EasingTypeOutQuint
	case "in_out_quint":
		return protocol.EasingTypeInOutQuint
	case "in_sine":
		return protocol.EasingTypeInSine
	case "out_sine":
		return protocol.EasingTypeOutSine
	case "in_out_sine":
		return protocol.EasingTypeInOutSine
	case "in_expo":
		return protocol.EasingTypeInExpo
	case "out_expo":
		return protocol.EasingTypeOutExpo
	case "in_out_expo":
		return protocol.EasingTypeInOutExpo
	case "in_circ":
		return protocol.EasingTypeInCirc
	case "out_circ":
		return protocol.EasingTypeOutCirc
	case "in_out_circ":
		return protocol.EasingTypeInOutCirc
	case "in_bounce":
		return protocol.EasingTypeInBounce
	case "out_bounce":
		return protocol.EasingTypeOutBounce
	case "in_out_bounce":
		return protocol.EasingTypeInOutBounce
	case "in_back":
		return protocol.EasingTypeInBack
	case "out_back":
		return protocol.EasingTypeOutBack
	case "in_out_back":
		return protocol.EasingTypeInOutBack
	case "in_elastic":
		return protocol.EasingTypeInElastic
	case "out_elastic":
		return protocol.EasingTypeOutElastic
	case "in_out_elastic":
		return protocol.EasingTypeInOutElastic
	}
	panic("shouldn't happen")
}

func CameraEaseString(i uint8) string {
	switch i {
	case protocol.EasingTypeLinear:
		return "linear"
	case protocol.EasingTypeSpring:
		return "spring"
	case protocol.EasingTypeInQuad:
		return "in_quad"
	case protocol.EasingTypeOutQuad:
		return "out_quad"
	case protocol.EasingTypeInOutQuad:
		return "in_out_quad"
	case protocol.EasingTypeInCubic:
		return "in_cubic"
	case protocol.EasingTypeOutCubic:
		return "out_cubic"
	case protocol.EasingTypeInOutCubic:
		return "in_out_cubic"
	case protocol.EasingTypeInQuart:
		return "in_quart"
	case protocol.EasingTypeOutQuart:
		return "out_quart"
	case protocol.EasingTypeInOutQuart:
		return "in_out_quart"
	case protocol.EasingTypeInQuint:
		return "in_quint"
	case protocol.EasingTypeOutQuint:
		return "out_quint"
	case protocol.EasingTypeInOutQuint:
		return "in_out_quint"
	case protocol.EasingTypeInSine:
		return "in_sine"
	case protocol.EasingTypeOutSine:
		return "out_sine"
	case protocol.EasingTypeInOutSine:
		return "in_out_sine"
	case protocol.EasingTypeInExpo:
		return "in_expo"
	case protocol.EasingTypeOutExpo:
		return "out_expo"
	case protocol.EasingTypeInOutExpo:
		return "in_out_expo"
	case protocol.EasingTypeInCirc:
		return "in_circ"
	case protocol.EasingTypeOutCirc:
		return "out_circ"
	case protocol.EasingTypeInOutCirc:
		return "in_out_circ"
	case protocol.EasingTypeInBounce:
		return "in_bounce"
	case protocol.EasingTypeOutBounce:
		return "out_bounce"
	case protocol.EasingTypeInOutBounce:
		return "in_out_bounce"
	case protocol.EasingTypeInBack:
		return "in_back"
	case protocol.EasingTypeOutBack:
		return "out_back"
	case protocol.EasingTypeInOutBack:
		return "in_out_back"
	case protocol.EasingTypeInElastic:
		return "in_elastic"
	case protocol.EasingTypeOutElastic:
		return "out_elastic"
	case protocol.EasingTypeInOutElastic:
		return "in_out_elastic"
	default:
		panic("shouldn't happen")
	}
}
