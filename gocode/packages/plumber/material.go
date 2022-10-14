/**
Some parts of sourcecode for the book written by Katherine Cox-Buday,
“Concurrency in Go” published by O’Reilly, are reused in this file with free adaptation.

Please read the copyright notice:
https://github.com/kat-co/concurrency-in-go-src/blob/master/LICENSE

*/

package plumber

import (
	"context"
	"sync"
)

func Regulated[IT any, OT any](
	ctx context.Context,
	rlatorName string,
	fn func(context.Context, IT) OT,
	input IT,
) (output OT) {
	config, _ := ConfigFromContext(ctx)
	select {
	case <-ctx.Done():
		return
	case <-config.rglatorChans[rlatorName]:
		defer func() {
			config.rglatorChans[rlatorName] <- struct{}{}
		}()
		output = fn(ctx, input)
	}
	return
}

func OrDone[T any](ctx context.Context, c chan T) chan T {
	valStream := make(chan T)
	go func() {
		defer close(valStream)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if ok == false {
					return
				}
				select {
				case valStream <- v:
				case <-ctx.Done():
				}
			}
		}
	}()
	return valStream
}

func Parallelize[IT any, OT any](
	ctx context.Context,
	rlatorName string,
	fn func(context.Context, IT) OT,
	inputs ...IT,
) (outputs []OT) {
	if len(inputs) == 0 {
		return
	}
	config, ok := ConfigFromContext(ctx)
	if !ok || !config.PlizerEnabled {
		for _, input := range inputs {
			output := fn(ctx, input)
			outputs = append(outputs, output)
		}
		return
	}

	outputStream := make(chan OT)
	var wg sync.WaitGroup

	plized := func(input IT) {
		defer wg.Done()
		var output OT
		if config.RglatorsByName[rlatorName] == 0 {
			output = fn(ctx, input)
		} else {
			output = Regulated[IT, OT](ctx, rlatorName, fn, input)
		}
		select {
		case <-ctx.Done():
			return
		case outputStream <- output:
		}
	}
	wg.Add(len(inputs))
	go func() {
		wg.Wait()
		close(outputStream)
	}()
	for _, input := range inputs {
		input := input
		go plized(input)
	}
	for output := range OrDone[OT](ctx, outputStream) {
		outputs = append(outputs, output)
	}
	return
}

type Launchable func(ctx context.Context, input interface{}) (output interface{})

func Retype[OT any](inputs []interface{}) (outputs []OT) {
	outputs = make([]OT, len(inputs))
	for i, input := range inputs {
		if input != nil {
			outputs[i] = input.(OT)
		}
	}
	return
}

func Untype[IT any](inputs []IT) (outputs []interface{}) {
	outputs = make([]interface{}, len(inputs))
	for i, input := range inputs {
		outputs[i] = input
	}
	return
}

func LaunchAndWait(
	ctx context.Context,
	rlatorNames []string,
	fns []Launchable,
	inputs []interface{},
) (outputs []interface{}) {
	config, ok := ConfigFromContext(ctx)
	if !ok || !config.PlizerEnabled {
		for i, fn := range fns {
			outputs = append(outputs, fn(ctx, inputs[i]))
		}
		return
	}

	outputs = make([]interface{}, len(fns))
	var wg sync.WaitGroup
	wg.Add(len(fns))

	for i, fn := range fns {
		i, fn := i, fn
		go func() {
			defer wg.Done()
			if config.RglatorsByName[rlatorNames[i]] == 0 {
				outputs[i] = fn(ctx, inputs[i])
			} else {
				outputs[i] = Regulated[interface{}, interface{}](ctx, rlatorNames[i], fn, inputs[i])
			}
		}()
	}

	wg.Wait()
	return
}
