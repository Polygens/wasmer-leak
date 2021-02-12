package main

import (
	"bufio"
	"io/ioutil"
	"math"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/wasmerio/wasmer-go/wasmer"
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

	engine := wasmer.NewEngine()
	store := wasmer.NewStore(engine)

	// Compiles the module
	module, err := wasmer.NewModule(store, wasmBytes)
	if err != nil {
		logrus.Fatalf("failed to compile webassembly binary: %s", err)
	}

	wasiEnv, err := wasmer.NewWasiStateBuilder("example").CaptureStdout().CaptureStderr().Finalize()
	if err != nil {
		logrus.Fatalf("failed to create env: %s", err)
	}

	importObject, err := wasiEnv.GenerateImportObject(store, module)
	if err != nil {
		logrus.Fatalf("failed to generate import object: %s", err)
	}

	instance, err := wasmer.NewInstance(module, importObject)
	if err != nil {
		logrus.Fatalf("Failed to instantiate module: %s", err)
	}

	memory, err := instance.Exports.GetMemory("memory")
	if err != nil {
		logrus.Fatalf("failed to get memory from instance: %s", err)
	}

	// Gets the `transform` exported function from the wasm instance.
	transform, err := instance.Exports.GetFunction("transform")
	if err != nil {
		// The wasm is missing `transform` and is therefore invalid
		logrus.Fatalf("transform function not found: %s", err)
	}

	for i := 0; i < 320000; i++ {
		data := memory.Data()
		copy(data, input)

		// Calls the exported function with the start and length of the inputJson in memory
		// The return value is an integer containing the start and length of the transformed json in memory
		output, err := transform(0, len(input))
		if err != nil {
			logrus.Fatal("failed to transform: %w", err)
		}

		// Retrieve integer from the return value
		memoryLocation, ok := output.(int64)
		if !ok {
			logrus.Fatalf("invalid output")
		}

		// Webassembly is limited to return only a single value but we need two integers:
		// 1. The pointer of the beginning of the string in memory
		// 2. The Length of the string
		// To work around this limitation we use cantor pairing which allows combining and extracting two integers inside a single integer.
		start, length := invertedCantorPairing(memoryLocation)

		// Retrieve the updated memory of the instance
		data = memory.Data()

		dataCopy := make([]byte, len(data))
		copy(dataCopy, data)

		logrus.Infof("Input JSON: %s", input)
		logrus.Infof("Output JSON: %s", dataCopy[start:start+length])
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
