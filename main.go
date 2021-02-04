package main

import (
	"bufio"
	"io/ioutil"
	"math"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wasmerio/go-ext-wasm/wasmer"
)

func main() {
	// Retrieve WASM binary
	wasmBytes, err := ioutil.ReadFile("./wasm.wasm")
	if err != nil {
		logrus.Fatalf("Failed to read wasm file: %s", err)
	}

	// Retrieve our sample input data
	input, err := ioutil.ReadFile("./input.json")
	if err != nil {
		logrus.Fatalf("Failed to read input file: %s", err)
	}

	wasmModule, err := wasmer.Compile(wasmBytes)
	if err != nil {
		logrus.Fatal(err)
	}
	defer wasmModule.Close()

	wasiVersion := wasmer.WasiGetVersion(wasmModule)
	wasmImportObject := wasmer.NewDefaultWasiImportObjectForVersion(wasiVersion)
	defer wasmImportObject.Close()

	for i := 0; i < 100; i++ {
		instance, err := wasmModule.InstantiateWithImportObject(wasmImportObject)
		if err != nil {
			logrus.Fatalf("Failed to instantiate module: %s", err)
		}

		data := instance.Memory.Data()
		copy(data, input)

		// Gets the `transform` exported function from the wasm instance.
		transform, ok := instance.Exports["transform"]
		if !ok {
			logrus.Fatal("transform function not found")
		}

		// Calls the exported function with the start and length of the inputJson in memory
		// The return value is an integer containing the start and length of the transformed json in memory
		output, err := transform(0, len(input))
		if err != nil {
			logrus.Fatal("failed to transform: %w", err)
		}

		// Retrieve integer from the return value
		memoryLocation := output.ToI64()

		// Webassembly is limited to return only a single value but we need two integers:
		// 1. The pointer of the beginning of the string in memory
		// 2. The Length of the string
		// To work around this limitation we use cantor pairing which allows combining and extracting two integers inside a single integer.
		start, length := invertedCantorPairing(memoryLocation)

		// Retrieve the updated memory of the instance
		data = instance.Memory.Data()

		dataCopy := make([]byte, len(data))
		copy(dataCopy, data)

		logrus.Infof("Input JSON: %s", input)
		logrus.Infof("Output JSON: %s", dataCopy[start:start+length])

		instance.Close()
		instance.Memory.Close()
	}

	// Wait for user input before exiting container
	cliInput := bufio.NewScanner(os.Stdin)
	cliInput.Scan()
}

func invertedCantorPairing(z int64) (int32, int32) {
	w := int64(math.Floor((math.Sqrt(float64(8*z+1)) - 1) / 2))
	t := (w*w + w) / 2
	y := z - t
	x := w - y
	return int32(x), int32(y)
}
