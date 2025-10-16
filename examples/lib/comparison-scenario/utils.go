package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/cli"
	"ocm.software/ocm/api/credentials"
	"ocm.software/ocm/api/ocm"
	"ocm.software/ocm/api/ocm/extensions/attrs/signingattr"
	"ocm.software/ocm/api/tech/signing/handlers/rsa"
)

func PrintPublicKey(ctx ocm.Context, name string) {
	info := signingattr.Get(ctx)
	key := info.GetPublicKey(name)
	if key == nil {
		fmt.Printf("public key for %s not found\n", name)
	} else {
		buf := bytes.NewBuffer(nil)
		err := rsa.WriteKeyData(key, buf)
		if err != nil {
			fmt.Printf("key error: %s\n", err)
		} else {
			fmt.Printf("public key for %s:\n%s\n", name, buf.String())
		}
	}
}

func PrintSignatures(cv ocm.ComponentVersionAccess) {
	fmt.Printf("signatures:\n")
	for i, s := range cv.GetDescriptor().Signatures {
		fmt.Printf("%2d    name: %s\n", i, s.Name)
		fmt.Printf("      digest:\n")
		fmt.Printf("        algorithm:     %s\n", s.Digest.HashAlgorithm)
		fmt.Printf("        normalization: %s\n", s.Digest.NormalisationAlgorithm)
		fmt.Printf("        value:         %s\n", s.Digest.Value)
		fmt.Printf("      signature:\n")
		fmt.Printf("        algorithm: %s\n", s.Signature.Algorithm)
		fmt.Printf("        mediaType: %s\n", s.Signature.MediaType)
		fmt.Printf("        value:     %s\n", s.Signature.Value)
	}
}

func PrintConsumerId(o interface{}, msg string) {
	// register credentials for given OCI registry in context.
	id := credentials.GetProvidedConsumerId(o)
	if id == nil {
		fmt.Printf("no consumer id for %s\n", msg)
	} else {
		fmt.Printf("consumer id for %s: %s\n", msg, id)
	}
}

func InstallChart(chart *chart.Chart, release, namespace string) error {
	settings := cli.New()
	settings.SetNamespace(namespace)
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(
		settings.RESTClientGetter(),
		namespace,
		os.Getenv("HELM_DRIVER"),
		func(msg string, args ...interface{}) { fmt.Printf(msg, args...) },
	); err != nil {
		return err
	}

	client := action.NewInstall(actionConfig)
	client.ReleaseName = release
	client.Namespace = namespace
	if _, err := client.Run(chart, nil); err != nil {
		return err
	}

	return nil
}

func Close(c io.Closer) {
	err := c.Close()
	if err != nil {
		panic(err)
	}
}
