[![Coverage Status](https://coveralls.io/repos/github/bjartek/underflow/badge.svg?branch=main)](https://coveralls.io/github/bjartek/underflow?branch=main) [![ci](https://github.com/bjartek/underflow/actions/workflows/ci.yaml/badge.svg)](https://github.com/bjartek/underflow/actions/workflows/ci.yaml)
# Underflow

A library extracted out from https://github.com/bjartek/overflow that has less depdendencies and only parses to from go<>cadence


## How to create a cadence value from a struct


```cadence
pub contract MyFancyContract {

	pub struct MyStruct{
		pub let owner: Address
		pub let name: String

}
````

Create the following go struct

```go

type MyFancyContract_MyStruct struct {
	Owner string `cadence:"owner,cadenceAddress"`
	Name string `cadence:"name"`
}


myImpl := MyFancyContract_MyStruct{
  Owner: "0x123",
  Name: "bjartek is the best"
}


//resolver is here a function that takes in the name of a go struct and returns the identifier of the cadence type on a given network
// resolver func(name string) (string, error) 

err, myCadenceValue :=underflow.InputToCadence(myImp, resolver)

```
