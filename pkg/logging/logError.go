package logging

func Error(err error) {
	if err != nil {
		println(err.Error())
	}
}
