package gp

import (
	"log"
	"math/rand"
)

//PopulationServer handle the population data in safe way
func PopulationServer(pop *Population) {
	for {
		msg := <-pop.Channel
		switch {
		case msg.Command == GET:
			idx := rand.Intn(len(pop.Individuos) - 1)
			ind := pop.Individuos[idx]

			a := make([]*Individuo, len(pop.Individuos)-1)
			a = append(pop.Individuos[:idx], pop.Individuos[idx+1:]...)
			pop.Individuos = a

			*msg.Channel <- *ind

		case msg.Command == PUT:
			pop.Individuos = append(pop.Individuos, msg.Individuo)
			*msg.Channel <- Individuo{}

		default:
			log.Println("Unknow message")
		}
	}
}
