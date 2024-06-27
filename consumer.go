package main

type Consumer interface {
	Accept(*EventData)

	BatchAccept([]*EventData)
}
