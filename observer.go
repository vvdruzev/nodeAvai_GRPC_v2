package main


type Observer interface {
	Notify(string)
}

func (p *publish) AddObserver(o Observer)  {
	p.ObserverList = append(p.ObserverList, o)
}

func (p *publish) RemoveObserver(o Observer)  {
	var indexToRemove  int

	for i, observer := range p.ObserverList {
		if observer == o {
			indexToRemove = i
			break
		}
	}

	p.ObserverList = append(p.ObserverList[:indexToRemove], p.ObserverList[indexToRemove+1:]...)
}

func (p *publish) NotyfiObserver(m string)  {
	for _,v := range p.ObserverList{
		v.Notify(m)
	}
}

type publish struct {
	ObserverList []Observer
}

func (p *publish) sendEvent (s string,ch chan string)  {
	ch <- s
}

type publisher interface {
	sendEvent(string, chan string)
	AddObserver(o Observer)
	RemoveObserver(o Observer)
	NotyfiObserver(m string)
}

type emptyPublish struct {

}

func (p *emptyPublish) AddObserver(o Observer) {
}

func (p *emptyPublish) RemoveObserver(o Observer) {
}

func (p *emptyPublish) NotyfiObserver(m string) {
}

func (p *emptyPublish) sendEvent (s string,ch chan string) {
}