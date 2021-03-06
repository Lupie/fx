// Copyright (c) 2016 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package config is the Configuration Package.
//
// At a high level, configuration is any data that is used in an application but
// not part of the application itself. Any reasonably complex system needs to
// have knobs to tune, and not everything can have intelligent defaults.
//
//
// In UberFx, we try very hard to make configuring UberFx convenient. Users can:
//
// • Get components working with minimal configuration
//
// • Override any field if the default doesn't make sense for their use case
//
// Nesting
//
// The configuration system wraps a set of *providers* that each know how to get
// values from an underlying source:
//
//
// • Static YAML configuration
//
// • Environment variables
//
// So by stacking these providers, we can have a priority system for defining
// configuration that can be overridden by higher priority providers. For example,
// the static YAML configuration would be the lowest priority and those values
// should be overridden by values specified as environment variables.
//
//
// As an example, imagine a YAML config that looks like:
//
//   foo:
//     bar:
//       boo: 1
//       baz: hello
//
//   stuff:
//     server:
//       port: 8081
//       greeting: Hello There!
//
// UberFx Config allows direct key access, such as foo.bar.baz:
//
//   cfg := svc.Config()
//   if value := cfg.Get("foo.bar.baz"); value.HasValue() {
//     fmt.Printf("Say %s", value.AsString()) // "Say hello"
//   }
//
// Or via a strongly typed structure, even as a nest value, such as:
//
//   type myStuff struct {
//     Port     int    `yaml:"port" default:"8080"`
//     Greeting string `yaml:"greeting"`
//   }
//
//   // ....
//
//   target := &myStuff{}
//   cfg := svc.Config()
//   if err := cfg.Get("stuff.server").PopulateStruct(target); err != nil {
//     // fail, we didn't find it.
//   }
//
//   fmt.Printf("Port is: %v", target.Port)
//
// Prints **Port is 8081**
//
// This model respects priority of providers to allow overriding of individual
// values.  In this example, we override the server port via an environment
// variable:
//
//
//   export CONFIG__stuff__server__port=3000
//
// Then running the above example will result in **Port is 3000**
//
// Provider
//
// Provider is the interface for anything that can provide values.
// We provide a few reference implementations (environment and YAML), but you are
// free to register your own providers via
// config.RegisterProviders() and
// config.RegisterDynamicProviders.
//
// Static configuration providers
//
// Static configuration providers conform to the Provider interface
// and are bootstraped first. Use these for simple providers such as file-backed or
// environment-based configuration providers.
//
//
// Dynamic Configuration Providers
//
// Dynamic configuration providers frequently need some bootstrap configuration to
// be useful, so UberFx treats them specially. Dynamic configuration providers
// conform to the
// Provider interface, but they're instantiated
// **after** the Static Providers on order to read bootstrap values.
//
// For example, if you were to implement a ZooKeeper-backed
// Provider, you'd likely need to specify (via YAML or environment
// variables) where your ZooKeeper nodes live.
//
//
// Value
//
// Value is the return type of every configuration providers'
// Get(key string) method. Under the hood, we use the empty interface
// (
// interface{}) since we don't necessarily know the structure of your
// configuration ahead of time.
//
//
// You can use a Value for two main purposes:
//
// • Get a single value out of configuration.
//
// For example, if we have a YAML configuration like so:
//
//   one:
//     two: hello
//
// You could access the value using "dotted notation":
//
//   foo := provider.Get("one.two").AsString()
//   fmt.Println(foo)
//   // Output: hello
//
// To get an access to the root element use config.Root:
//
//   root := provider.Get(config.Root).AsString()
//   fmt.Println(root)
//   // Output: map[one:map[two:hello]]
//
// • Populate a struct (PopulateStruct(&myStruct))
//
// The As* method has two variants: TryAs* and As*. The former is a
// two-value return, similar to a type assertion, where the user checks if the second
// bool is true before using the value.
//
// The As* methods are similar to the Must* pattern in the standard library.
// If the underlying value cannot be converted to the requested type,
// As* will
// panic.
//
// PopulateStruct
//
// PopulateStruct is akin to json.Unmarshal() in that it takes a pointer to a
// custom struct and fills in the fields. It returns a
// true if the requested
// fields were found and populated properly, and
// false otherwise.
//
// For example, say we have the following YAML file:
//
//   hello:
//     world: yes
//     number: 42
//
// We could deserialize into our custom type with the following code:
//
//   type myConfig struct {
//     World  string
//     Number int
//   }
//
//   m := myConfig{}
//   provider.Get("hello").Populate(&m)
//   fmt.Println(m.World)
//   // Output: yes
//
// Note that any fields you wish to deserialize into must be exported, just like
// json.Unmarshal and friends.
//
// Benchmarks
//
// Current performance benchmark data:
//
//   BenchmarkYAMLCreateSingleFile-8                       50000     31269 ns/op   11136 B/op     121 allocs/op
//   BenchmarkYAMLCreateMultiFile-8                        30000     52378 ns/op   20064 B/op     205 allocs/op
//   BenchmarkYAMLSimpleGetLevel1-8                     50000000      27.1 ns/op       0 B/op       0 allocs/op
//   BenchmarkYAMLSimpleGetLevel3-8                     50000000      26.8 ns/op       0 B/op       0 allocs/op
//   BenchmarkYAMLSimpleGetLevel7-8                     50000000      26.3 ns/op       0 B/op       0 allocs/op
//   BenchmarkYAMLPopulateStruct-8                       2000000       861 ns/op     192 B/op      10 allocs/op
//   BenchmarkYAMLPopulateStructNested-8                  500000      2616 ns/op     616 B/op      34 allocs/op
//   BenchmarkYAMLPopulateStructNestedMultipleFiles-8     500000      3330 ns/op     744 B/op      42 allocs/op
//   BenchmarkYAMLPopulateNestedTextUnmarshaler-8         100000     16775 ns/op    3201 B/op     209 allocs/op
//
//
package config
