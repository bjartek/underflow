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


myImpl := MyFancyContract{
  Owner: "0x123",
  Name: "Asd"
}


//resolver is here a function that takes in the name of a struct and returns the identifier on the given network. 

// resolver func(name string) (string, error) input is name of struct, return value is identifier on the current network
err, myCadenceValue :=underflow.InputToCadence(myImp, resolver)

```
