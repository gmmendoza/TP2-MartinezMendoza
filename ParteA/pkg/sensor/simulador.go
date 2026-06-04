package sensor

import (
	"math/rand"
	"sync"
)

// Simulador genera lecturas de temperatura de forma thread-safe.
type Simulador struct {
	mu            sync.RWMutex
	ultimaLectura float64
}

// NuevoSimulador crea un simulador con una lectura inicial.
func NuevoSimulador() *Simulador {
	return &Simulador{
		ultimaLectura: 22.0 + rand.Float64()*5.0, // entre 22.0 y 27.0
	}
}

// Leer devuelve una nueva lectura simulada y la almacena.
func (s *Simulador) Leer() float64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	variacion := rand.Float64() - 0.5

	nuevaTemperatura := s.ultimaLectura + variacion

	if nuevaTemperatura < 15.0 {
		nuevaTemperatura = 15.0
	} else if nuevaTemperatura > 35.0 {
		nuevaTemperatura = 35.0
	}

	s.ultimaLectura = nuevaTemperatura
	return s.ultimaLectura
}

// ObtenerUltima devuelve la última lectura sin generar una nueva.
func (s *Simulador) ObtenerUltima() float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ultimaLectura
}
