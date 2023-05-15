package types

import (
	// These packages are imported for side effects!
	// They have to be imported so that their init-functions are executed and their respective types are added
	// (or rather registered) to the registy.DefaultMessageRegistry.
	_ "example/prototype-based-typeregistry/types/complexmessage"
	_ "example/prototype-based-typeregistry/types/simplemessage"
)
