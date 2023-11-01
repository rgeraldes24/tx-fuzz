package spammer

import (
	"fmt"
	"math/rand"
	"sync"

	"github.com/MariusVanDerWijden/FuzzyVM/filler"
	"github.com/theQRL/go-qrllib/dilithium"
)

type Spam func(*Config, *dilithium.Dilithium, *filler.Filler) error

func SpamTransactions(config *Config, fun Spam) error {
	fmt.Printf("Spamming %v transactions per account on %v accounts with seed: 0x%x\n", config.N, len(config.accs), config.seed)

	errCh := make(chan error, len(config.accs))
	var wg sync.WaitGroup
	wg.Add(len(config.accs))
	for _, acc := range config.accs {
		// Setup randomness uniquely per key
		random := make([]byte, 10000)
		config.mut.FillBytes(&random)

		var f *filler.Filler
		if len(config.corpus) != 0 {
			elem := config.corpus[rand.Int31n(int32(len(config.corpus)))]
			config.mut.MutateBytes(&elem)
			f = filler.NewFiller(elem)
		} else {
			// Use lower entropy randomness for filler
			config.mut.MutateBytes(&random)
			f = filler.NewFiller(random)
		}
		// Start a fuzzing thread
		go func(acc *dilithium.Dilithium, filler *filler.Filler) {
			defer wg.Done()
			errCh <- fun(config, acc, f)
		}(acc, f)
	}
	wg.Wait()
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}
