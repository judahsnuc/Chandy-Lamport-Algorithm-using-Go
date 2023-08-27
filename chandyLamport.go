package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Process struct {
	id int
	bankAccounts [3]int
	incoming chan Message
	outgoing []chan Message
	localSnapshot int
	channelState []int
	mu sync.Mutex
}

type Message struct {
	sender int
	value int
	marker bool
}

func NewProcess(id int, outgoing []chan Message) *Process {
	p := &Process{
		id: id,
		bankAccounts: [3]int{1000, 1000, 1000},
		incoming: make(chan Message, 100),
		outgoing: outgoing,
		localSnapshot: 3000,
		channelState: make([]int, len(outgoing)),
	}
	go p.run()
	return p
}

func (p *Process) run() {
	for m := range p.incoming {
		if m.marker {
			p.takeSnapshot(m.sender)
		} else {
			p.bankAccounts[0] += m.value
			p.localSnapshot += m.value
			p.channelState[m.sender] += m.value
		}
	}
}

func (p *Process) takeSnapshot(sender int) {
	p.mu.Lock()
	defer p.mu.Unlock()
	fmt.Printf("Process %d taking snapshot\n", p.id)
	fmt.Printf("Local state: %d\n", p.localSnapshot)
	fmt.Printf("Channel states: %v\n", p.channelState)
	for i := range p.outgoing {
		if i != sender {
			p.outgoing[i] <- Message{sender: p.id, marker: true}
		}
	}
}

func (p *Process) transaction(n int) {
	for i := 0; i < n*15; i++ {
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
		target := rand.Intn(len(p.outgoing))
		value := rand.Intn(100)
		p.outgoing[target] <- Message{sender: p.id, value: value}
		p.bankAccounts[1] -= value
		p.localSnapshot -= value
		if i % n == 0 {
			p.takeSnapshot(-1)
			fmt.Printf("Total amount in snapshot: %d\n", p.localSnapshot + sum(p.channelState))
		}
	}
}

func sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

func main() {
	var n int
	fmt.Print("Enter the number of processes: ")
	fmt.Scan(&n)

	processes := make([]*Process, n)
	channels := make([][]chan Message, n)
	for i := range channels {
		channels[i] = make([]chan Message, n)
		for j := range channels[i] {
			channels[i][j] = make(chan Message, 100)
		}
	}

	for i := range processes {
		processes[i] = NewProcess(i, channels[i])
		for j := range channels {
			if i != j {
				processes[i].outgoing = append(processes[i].outgoing, channels[j][i])
			}
		}
	}

	fmt.Println("Total amount in system:", n * 3000)

	var wg sync.WaitGroup
	wg.Add(n)
	for _, p := range processes {
		go func(p *Process) {
			defer wg.Done()
			p.transaction(n)
			close(p.incoming)
			fmt.Printf("Process %d finished\n", p.id)
			fmt.Printf("Bank accounts: %v\n", p.bankAccounts)
			fmt.Printf("Local state: %d\n", p.localSnapshot)
			fmt.Printf("Channel states: %v\n", p.channelState)
			fmt.Printf("Total amount in snapshot: %d\n", p.localSnapshot + sum(p.channelState))
			
			if (p.localSnapshot + sum(p.channelState)) != (n * 3000) {
				panic("Incorrect snapshot")
			}
			
			for _, c := range p.outgoing {
				close(c)
			}
			
			wg.Done()
			
        }(p)
    }
    
    wg.Wait()
    
    fmt.Println("All processes finished")
    
    totalAmountInSystem := 0
    
    for _, p := range processes{
        totalAmountInSystem += sum(p.bankAccounts[:])
    }
    
    fmt.Println("Total amount in system:", totalAmountInSystem)
    
}
