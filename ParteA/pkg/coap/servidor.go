package coap

import (
	"bytes"
	"encoding/json"
	"log"
	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"
	"sync"
	"time"

	"github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/mux"
)

// ServidorCoAP expone recursos REST sobre UDP.
type ServidorCoAP struct {
	sim    *sensor.Simulador
	config nodo.Configuracion
	mu     sync.RWMutex
	modo   string
}

// NuevoServidor crea la instancia del servidor CoAP.
func NuevoServidor(sim *sensor.Simulador, config nodo.Configuracion) *ServidorCoAP {
	return &ServidorCoAP{
		sim:    sim,
		config: config,
		modo:   "automatico",
	}
}

// Iniciar arranca el servidor UDP en el puerto 5683.
func (s *ServidorCoAP) Iniciar() {
	// 6a. Crear router con mux.NewRouter()
	r := mux.NewRouter()

	// 6b. Registrar handler GET /temperatura
	r.HandleFunc("/temperatura", func(w mux.ResponseWriter, req *mux.Message) {
		if req.Code() != codes.GET {
			w.SetResponse(codes.MethodNotAllowed, message.TextPlain, bytes.NewReader([]byte("Método no permitido")))
			return
		}

		// Estructura para  la salida esperada en formato JSON
		type RespuestaTemperatura struct {
			NodoID      string    `json:"nodo_id"`
			Temperatura float64   `json:"temperatura"`
			Unidad      string    `json:"unidad"`
			Timestamp   time.Time `json:"timestamp"`
		}

		tempActual := s.sim.ObtenerUltima()

		resp := RespuestaTemperatura{
			NodoID:      s.config.ID,
			Temperatura: tempActual,
			Unidad:      "Celsius",
			Timestamp:   time.Now(),
		}

		payload, err := json.Marshal(resp)
		if err != nil {
			log.Printf("[CoAP-Error] Error serializando temperatura: %v", err)
			w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte("Internal Server Error")))
			return
		}

		w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader(payload))
	})

	// 6c. Registrar handler PUT /config y GET /config
	r.HandleFunc("/config", func(w mux.ResponseWriter, req *mux.Message) {

		switch req.Code() {
		case codes.PUT:
			body, err := req.ReadBody()
			if err != nil {
				w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte(" Error body invalido")))
				return
			}

			var cambios map[string]interface{}
			if err := json.Unmarshal(body, &cambios); err != nil {
				w.SetResponse(codes.BadRequest, message.TextPlain, bytes.NewReader([]byte("JSON invalido")))
				return
			}

			s.mu.Lock()
			if nuevoModo, ok := cambios["modo"].(string); ok {
				s.modo = nuevoModo
			}
			s.mu.Unlock()

			w.SetResponse(codes.Changed, message.TextPlain, bytes.NewReader([]byte("Configuracion actualizada con exito")))

		case codes.GET:
			// 6d. Respuesta del estado actual si entran por GET /config
			s.mu.RLock()
			configActual := map[string]interface{}{
				"modo":           s.modo,
				"config_inicial": s.config,
			}
			s.mu.RUnlock()

			payload, err := json.Marshal(configActual)
			if err != nil {
				w.SetResponse(codes.InternalServerError, message.TextPlain, bytes.NewReader([]byte("Error interno")))
				return
			}
			w.SetResponse(codes.Content, message.AppJSON, bytes.NewReader(payload))

		default:
			w.SetResponse(codes.MethodNotAllowed, message.TextPlain, bytes.NewReader([]byte("Metodo no permitido")))
		}
	})

	// 6e. Llamar coap.ListenAndServe("udp", ":5683", router)
	log.Println("[CoAP] Escuchando activamente en puerto UDP :5683...")
	if err := coap.ListenAndServe("udp", ":5683", r); err != nil {
		log.Fatalf("[CoAP-Fatal] El servidor falló catastróficamente: %v", err)
	}
}
