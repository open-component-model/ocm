package logging

func Error(err error, keypairs ...interface{}) {
	if err != nil {
		Context().Logger().Error(err.Error(), keypairs...)
	}
}
