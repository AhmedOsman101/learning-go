package exercises

type Temp = uint8

const (
	Celsius Temp = iota
	Fahrenheit
	Kelvin
)

func TempConvertor(temp float64, mode Temp) (celsius, fahrenheit, kelvin float64) {
	nineOverFive := (float64(9) / float64(5))
	fiveOverNine := (float64(5) / float64(9))
	switch mode {
	case Celsius:
		celsius = temp
		kelvin = celsius + 273.15
		fahrenheit = (celsius * nineOverFive) + 32
	case Fahrenheit:
		fahrenheit = temp
		celsius = (fahrenheit - 32) * fiveOverNine
		kelvin = ((fahrenheit - 32) * fiveOverNine) + 273.15
	case Kelvin:
		kelvin = temp
		celsius = kelvin - 273.15
		fahrenheit = ((kelvin - 273.15) * nineOverFive) + 32
	default:
		return 0, 0, 0
	}
	return
}
