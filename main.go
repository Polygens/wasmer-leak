package main

import (
	"io/ioutil"

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

	data := memory.Data()
	copy(data, input)

	for {
		// Calls the exported function with the start and length of the inputJson in memory
		// The return value is an integer containing the start and length of the transformed json in memory
		_, err := transform(0, len(input))
		if err != nil {
			logrus.Fatal("failed to transform: %w", err)
		}
	}
}
