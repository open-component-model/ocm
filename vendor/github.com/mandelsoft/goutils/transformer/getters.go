package transformer

// typical getters

func GetKey[E interface{ GetKey() R }, R any](e E) R {
	return e.GetKey()
}

func GetName[E interface{ GetName() R }, R any](e E) R {
	return e.GetName()
}

func GetDescription[E interface{ GetDescription() R }, R any](e E) R {
	return e.GetDescription()
}

func GetVersionn[E interface{ GetVersion() R }, R any](e E) R {
	return e.GetVersion()
}
