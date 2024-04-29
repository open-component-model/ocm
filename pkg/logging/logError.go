package logging

func Error(err error, keypairs ...interface{}) {
	if err != nil {
		Context().Logger(DefineSubRealm("error")).Error(err.Error(), keypairs...)
	}
}
