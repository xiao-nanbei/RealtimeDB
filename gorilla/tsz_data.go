package gorilla


type Point struct {
	V float64
	T uint64
}

// 120 points every 60s
var TwoHoursData = []Point{
	{761, 1440583200111}, {727, 1440583260222}, {765, 1440583320333}, {706, 1440583380444}, {700, 1440583440555},
}