package trade

import (
	"reflect"
	"testing"
	"time"
)

func (engine *Engine) expect(t *testing.T, ch chan Reply, expected []int64) (Reply, error) {
	for {
		select {
		case <-time.After(engine.timeout):
			return nil, timeoutError()
		case v := <-ch:
			if v.code() == 0 {
				t.Fatalf("don't know message '%v'", v)
			}
			for _, code := range expected {
				if v.code() == code {
					return v, nil
				}
			}
			// wrong message received
			t.Logf("received message '%v' of type '%v'\n",
				v, reflect.ValueOf(v).Type())
		}
	}

	return nil, nil
}

func TestConnect(t *testing.T) {
	engine, err := NewEngine()

	if err != nil {
		t.Fatalf("cannot connect engine: %s", err)
	}

	engine.Stop()
}

type Sink struct {
	ch chan Reply
}

func (self *Sink) Observe(v Reply) {
	self.ch <- v
}

func TestMarketData(t *testing.T) {
	engine, err := NewEngine()

	if err != nil {
		t.Fatalf("cannot connect engine: %s", err)
	}

	defer engine.Stop()

	req1 := &RequestMarketData{
		Contract: Contract{
			Symbol:       "AAPL",
			SecurityType: "STK",
			Exchange:     "SMART",
			Currency:     "USD",
		},
	}

	id := engine.NextRequestId()
	req1.SetId(id)
	ch := make(chan Reply)
	engine.Subscribe(&Sink{ch}, id)

	if err := engine.Send(req1); err != nil {
		t.Fatalf("client %d: cannot send market data request: %s", engine.ClientId(), err)
	}

	rep1, err := engine.expect(t, ch, []int64{mTickPrice, mTickSize})

	if err != nil {
		t.Fatalf("client %d: cannot receive market data: %s", engine.ClientId(), err)
	}

	t.Logf("received packet '%v' of type %v\n", rep1, reflect.ValueOf(rep1).Type())

	if err := engine.Send(&CancelMarketData{id}); err != nil {
		t.Fatalf("client %d: cannot send cancel request: %s", engine.ClientId(), err)
	}
}

func TestContractDetails(t *testing.T) {
	engine, err := NewEngine()

	if err != nil {
		t.Fatalf("cannot connect engine: %s", err)
	}

	defer engine.Stop()

	req1 := &RequestContractData{
		Contract: Contract{
			Symbol:       "AAPL",
			SecurityType: "STK",
			Exchange:     "SMART",
			Currency:     "USD",
		},
	}

	id := engine.NextRequestId()
	req1.SetId(id)
	ch := make(chan Reply)
	engine.Subscribe(&Sink{ch}, id)
	defer engine.Unsubscribe(id)

	if err := engine.Send(req1); err != nil {
		t.Fatalf("client %d: cannot send contract data request: %s", engine.ClientId(), err)
	}

	rep1, err := engine.expect(t, ch, []int64{mContractData})

	if err != nil {
		t.Fatalf("client %d: cannot receive contract details: %s", engine.ClientId(), err)
	}

	t.Logf("received packet '%v' of type %v\n", rep1, reflect.ValueOf(rep1).Type())

	rep2, err := engine.expect(t, ch, []int64{mContractDataEnd})

	if err != nil {
		t.Fatalf("client %d: cannot receive end of contract details: %s", engine.ClientId(), err)
	}

	t.Logf("received packet '%v' of type %v\n", rep2, reflect.ValueOf(rep2).Type())
}

func TestOptionChainRequest(t *testing.T) {
	engine, err := NewEngine()

	if err != nil {
		t.Fatalf("cannot connect engine: %s", err)
	}

	defer engine.Stop()

	req1 := &RequestContractData{
		Contract: Contract{
			Symbol:       "AAPL",
			SecurityType: "OPT",
			Exchange:     "SMART",
			Currency:     "USD",
		},
	}

	id := engine.NextRequestId()
	req1.SetId(id)
	ch := make(chan Reply)
	engine.Subscribe(&Sink{ch}, id)
	defer engine.Unsubscribe(id)

	if err := engine.Send(req1); err != nil {
		t.Fatalf("cannot send contract data request: %s", err)
	}

	rep1, err := engine.expect(t, ch, []int64{mContractDataEnd})

	if err != nil {
		t.Fatalf("cannot receive contract details: %s", err)
	}

	t.Logf("received packet '%v' of type %v\n", rep1, reflect.ValueOf(rep1).Type())
}
