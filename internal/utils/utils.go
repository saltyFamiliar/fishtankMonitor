package utils

func Must(action string, err error) {
	if err != nil {
		println("couldn't ", action)
		panic(err)
	}
}
