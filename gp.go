package gp

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"
)

//Individuo is the genetic code of solution
type Individuo struct {
	Genes   []int
	Fitness float64
}

//NewIndividuoRand create a new individuo with randon values
func NewIndividuoRand(length int) Individuo {
	ind := Individuo{}
	ind.Genes = make([]int, length)
	for i := 0; i < length; i++ {
		ind.Genes[i] = rand.Int()
	}
	return ind
}

//NewIndividuo create a new individuo
func NewIndividuo(length int) Individuo {
	ind := Individuo{}
	ind.Genes = make([]int, length)
	return ind
}

//String return a string representation of Individuo
func (ind Individuo) String() string {
	str := fmt.Sprintf("(%.2f,%d)", ind.Fitness, len(ind.Genes))
	str += "->["
	for _, val := range ind.Genes {
		str += strconv.Itoa(val)
		str += " "
	}
	return str + "]"
}

const (
	//GET is a tag message
	GET = iota

	//PUT is a tag message
	PUT
)

const (
	//GP Genetic Programming
	GP = iota

	//GA Genetic Algorithm
	GA
)

type populationMessage struct {
	Command   int
	Channel   *chan Individuo
	Individuo *Individuo
}

//Population represents an individuos population
type Population struct {
	TxCross      float64
	TxMut        float64
	Individuos   []*Individuo
	Channel      chan populationMessage      `json:"-"`
	FunFitness   func(ind Individuo) float64 `json:"-"`
	BoltSerie    string
	CountFitness int
	BestFit      float64
	BestHistory  map[string]Individuo
	Verbose      bool
	TypeAlg      int
}

//NewPopulation create a new random population
func NewPopulation(t int, tind int, txc float64, txm float64, f func(Individuo) float64) *Population {
	pop := Population{}
	pop.TxCross = txc
	pop.TxMut = txm
	pop.FunFitness = f
	pop.BestFit = 0
	pop.BoltSerie = "gp"
	pop.CountFitness = 0
	pop.TypeAlg = GP
	pop.BestHistory = map[string]Individuo{}
	pop.Channel = make(chan populationMessage)

	for i := 0; i < t; i++ {
		ind := NewIndividuoRand(tind)
		pop.Individuos = append(pop.Individuos, &ind)
	}

	go PopulationServer(&pop)
	return &pop
}

func (pop *Population) String() string {
	str := "{\n"
	for _, val := range pop.Individuos {
		str += val.String()
		str += "\n"
	}
	return str + "}"
}

//GetIndividuo in safe way
func (pop *Population) GetIndividuo() *Individuo {
	c := make(chan Individuo)
	msg := populationMessage{
		Command: GET,
		Channel: &c}
	pop.Channel <- msg
	ind := <-c
	return &ind
}

//PutIndividuo in a safe way
func (pop *Population) PutIndividuo(ind *Individuo) {
	c := make(chan Individuo)
	msg := populationMessage{
		Command:   PUT,
		Channel:   &c,
		Individuo: ind}
	pop.Channel <- msg
	<-c
}

//CrossoverGP apply the crossover operator to individuos
func CrossoverGP(ind1, ind2 *Individuo, txCross float64) (rind1, rind2 *Individuo) {
	r := rand.Float64()
	if r < txCross {
		//TODO ver a possibilidade de colocar uma normal para corte
		cut1 := rand.Intn(len(ind1.Genes) - 1)
		cut2 := rand.Intn(len(ind2.Genes) - 1)
		t1 := cut1 + len(ind2.Genes) - cut2
		t2 := cut2 + len(ind1.Genes) - cut1

		r1 := NewIndividuoRand(t1)
		r1.Genes = append(ind1.Genes[:cut1], ind2.Genes[cut2:]...)

		r2 := NewIndividuoRand(t2)
		r2.Genes = append(ind2.Genes[:cut2], ind1.Genes[cut1:]...)

		return &r1, &r2
	}

	r1 := Individuo{}
	r1.Genes = make([]int, len(ind1.Genes))
	copy(r1.Genes, ind1.Genes)
	r1.Fitness = ind1.Fitness

	r2 := Individuo{}
	r2.Genes = make([]int, len(ind2.Genes))
	copy(r2.Genes, ind2.Genes)
	r2.Fitness = ind2.Fitness
	return &r1, &r2
}

//CrossoverGA apply the crossover operator to individuos
func CrossoverGA(ind1, ind2 *Individuo, txCross float64) (rind1, rind2 *Individuo) {
	r := rand.Float64()
	if r < txCross {
		cut := rand.Intn(len(ind1.Genes) - 1)

		l := len(ind1.Genes)
		r1 := NewIndividuoRand(l)
		r1.Genes = append(ind1.Genes[:cut], ind2.Genes[cut:]...)

		r2 := NewIndividuoRand(l)
		r2.Genes = append(ind2.Genes[:cut], ind1.Genes[cut:]...)

		return &r1, &r2
	}
	r1 := Individuo{}
	r1.Genes = make([]int, len(ind1.Genes))
	copy(r1.Genes, ind1.Genes)
	r1.Fitness = ind1.Fitness

	r2 := Individuo{}
	r2.Genes = make([]int, len(ind2.Genes))
	copy(r2.Genes, ind2.Genes)
	r2.Fitness = ind2.Fitness
	return &r1, &r2
}

//Mutation implements the mutation operator
func Mutation(ind *Individuo, txmut float64) {
	for idx := range ind.Genes {
		r := rand.Float64()
		if r <= txmut {
			ind.Genes[idx] = rand.Int()
		}
	}
}

//Tournament implements the tournament strategy
func (pop *Population) Tournament() {
	ind1 := pop.GetIndividuo()
	ind2 := pop.GetIndividuo()
	ind3 := pop.GetIndividuo()
	ind4 := pop.GetIndividuo()

	var champion1, champion2 *Individuo

	if ind1.Fitness == 0.0 {
		ind1.Fitness = pop.FunFitness(*ind1)
		if pop.BestFit < ind1.Fitness {
			pop.BestHistory[fmt.Sprint(ind1.Fitness)] = *ind1
			pop.BestFit = ind1.Fitness
		}
	}
	if ind2.Fitness == 0.0 {
		ind2.Fitness = pop.FunFitness(*ind2)
		if pop.BestFit < ind2.Fitness {
			pop.BestHistory[fmt.Sprint(ind2.Fitness)] = *ind2
			pop.BestFit = ind2.Fitness
		}
	}
	if ind3.Fitness == 0.0 {
		ind3.Fitness = pop.FunFitness(*ind3)
		if pop.BestFit < ind3.Fitness {
			pop.BestHistory[fmt.Sprint(ind3.Fitness)] = *ind3
			pop.BestFit = ind3.Fitness
		}
	}
	if ind4.Fitness == 0.0 {
		ind4.Fitness = pop.FunFitness(*ind4)
		if pop.BestFit < ind4.Fitness {
			pop.BestHistory[fmt.Sprint(ind4.Fitness)] = *ind4
			pop.BestFit = ind4.Fitness
		}
	}

	if pop.Verbose {
		log.Printf("%14.2f(%4d),%14.2f(%4d),%14.2f(%4d),%14.2f(%4d),%14.2f\n",
			ind1.Fitness, len(ind1.Genes),
			ind2.Fitness, len(ind2.Genes),
			ind3.Fitness, len(ind3.Genes),
			ind4.Fitness, len(ind4.Genes),
			pop.BestFit)
	}

	if ind1.Fitness < ind2.Fitness {
		champion1 = ind2
	} else {
		champion1 = ind1
	}

	if ind3.Fitness < ind4.Fitness {
		champion2 = ind4
	} else {
		champion2 = ind3
	}

	Crossover := CrossoverGP
	if pop.TypeAlg == GA {
		Crossover = CrossoverGA
	}

	son1, son2 := Crossover(champion1, champion2, pop.TxCross)
	Mutation(son1, pop.TxMut)
	Mutation(son2, pop.TxMut)

	pop.PutIndividuo(champion1)
	pop.PutIndividuo(champion2)
	pop.PutIndividuo(son1)
	pop.PutIndividuo(son2)
}

func (pop *Population) Save(filename string) {
	r, err := json.MarshalIndent(*pop, " ", "   ")
	if err != nil {
		log.Fatalln("Marshal", err)
		return
	}

	err = ioutil.WriteFile(filename, r, 0644)
	if err != nil {
		log.Fatalln("writefile", err)
		return
	}
}
