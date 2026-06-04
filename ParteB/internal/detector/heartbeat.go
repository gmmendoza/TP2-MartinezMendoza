package detector

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"

	"sd-comunicacion/pkg/protocolo"
)

// Enviador se encarga de enviar heartbeats UDP periodicamente
type Enviador struct {
	destino   string
	intervalo time.Duration
	nodoID    string
	contador  int
}

// TODO 5: Implementar la funcion NuevoEnviador.
func NuevoEnviador(destino string, intervalo time.Duration, nodoID string) *Enviador {
	return &Enviador{
		destino:   destino,
		intervalo: intervalo,
		nodoID:    nodoID,
		contador:  0,
	}
}

// TODO 6: Implementar el metodo (e *Enviador) Iniciar().
func (e *Enviador) Iniciar() {
	log.Printf("[UDP-Enviador] Iniciando transmisión periódica hacia %s cada %v\n", e.destino, e.intervalo)

	ticker := time.NewTicker(e.intervalo)
	defer ticker.Stop()

	for range ticker.C {
		e.contador++

		hb := protocolo.Heartbeat{
			NodoID:    e.nodoID,
			Timestamp: time.Now().Unix(),
			Contador:  e.contador,
		}

		jsonData, err := json.Marshal(hb)
		if err != nil {
			log.Printf("[UDP-Enviador Error] Error serializando heartbeat: %v\n", err)
			continue
		}

		conn, err := net.Dial("udp", e.destino)
		if err != nil {
			log.Printf("[UDP-Enviador Error] No se pudo establecer socket UDP: %v\n", err)
			continue
		}

		_, err = conn.Write(jsonData)
		if err != nil {
			log.Printf("[UDP-Enviador Error] Falla al enviar: %v\n", err)
		}
		conn.Close()
	}
}

type Receptor struct {
	mu      sync.RWMutex
	puerto  string
	timeout time.Duration
	ultimo  time.Time
	estado  string
}

// TODO 7: Implementar la funcion NuevoReceptor.
func NuevoReceptor(puerto string, timeout time.Duration) *Receptor {
	return &Receptor{
		puerto:  puerto,
		timeout: timeout,
		ultimo:  time.Now(),
		estado:  "alive",
	}
}

// TODO 8: Implementar el metodo (r *Receptor) Escuchar().
func (r *Receptor) Escuchar() {

	packetConn, err := net.ListenPacket("udp", r.puerto)
	if err != nil {
		log.Fatalf("[UDP-Receptor Fatal] No se pudo abrir el puerto UDP %s: %v\n", r.puerto, err)
	}
	defer packetConn.Close()

	log.Printf("[UDP-Receptor] Escuchando heartbeats en %s...\n", r.puerto)

	go func() {
		tickerMonitoreo := time.NewTicker(1 * time.Second)
		defer tickerMonitoreo.Stop()

		for range tickerMonitoreo.C {
			r.mu.Lock()
			tiempoInactivo := time.Since(r.ultimo)
			estadoAnterior := r.estado

			if tiempoInactivo > 2*r.timeout {
				r.estado = "dead"
			} else if tiempoInactivo > r.timeout {
				r.estado = "suspect"
			} else {
				r.estado = "alive"
			}

			if r.estado != estadoAnterior {
				log.Printf("[DETECTOR DE FALLAS] Transición de estado detectada: %s -> %s (Inactividad: %v)\n",
					estadoAnterior, r.estado, tiempoInactivo.Round(time.Millisecond))
			}
			r.mu.Unlock()
		}
	}()

	buffer := make([]byte, 1024)
	for {
		n, _, err := packetConn.ReadFrom(buffer)
		if err != nil {
			log.Printf("[UDP-Receptor Error] Error leyendo del buffer UDP: %v\n", err)
			continue
		}

		var hb protocolo.Heartbeat
		if err := json.Unmarshal(buffer[:n], &hb); err != nil {
			log.Printf("[UDP-Receptor Peligro] Paquete corrupto o JSON inválido recibido: %v\n", err)
			continue
		}

		r.mu.Lock()
		r.ultimo = time.Now()

		if r.estado != "alive" {
			log.Printf(" [DETECTOR DE FALLAS] El nodo '%s' se ha recuperado. Estado: alive\n", hb.NodoID)
			r.estado = "alive"
		}
		r.mu.Unlock()
	}
}

func (r *Receptor) ObtenerEstado() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.estado
}
