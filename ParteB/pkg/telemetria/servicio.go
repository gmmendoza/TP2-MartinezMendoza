package telemetria

import (
	"fmt"
	"log"
	"sync"
	"time"

	"sd-comunicacion/pkg/protocolo"
)

// TODO 1: Definir el struct Telemetria que sera el servicio RPC.
type Telemetria struct {
	mu         sync.Mutex
	lecturas   map[string]protocolo.Lectura
	contadorID int64
}

func NuevoServicio() *Telemetria {
	return &Telemetria{
		lecturas: make(map[string]protocolo.Lectura),
	}
}

// TODO 2: Implementar el metodo RPC RegistrarLectura.
// Firma requerida por net/rpc: func (t *Telemetria) RegistrarLectura(args protocolo.Lectura, resp *protocolo.RespuestaLectura) error
func (t *Telemetria) RegistrarLectura(args protocolo.Lectura, resp *protocolo.RespuestaLectura) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Guardar la lectura en el mapa protegido
	t.lecturas[args.SensorID] = args

	t.contadorID++
	resp.ID = t.contadorID
	resp.Mensaje = "Lectura registrada con éxito" // Se adapta al nuevo campo string 'mensaje' del PDF

	fechaLegible := time.Unix(args.Timestamp, 0).Format("2006-01-02 15:04:05")
	log.Printf("[RPC-Server] Lectura registrada exitosamente [#%d] -> Sensor: %s | Temp: %.2f°C | Orig-TS: %s\n",
		resp.ID, args.SensorID, args.Temperatura, fechaLegible)

	return nil
}

// TODO 3: Implementar el metodo RPC ObtenerUltimaLectura.
// Firma requerida por net/rpc: func (t *Telemetria) ObtenerUltimaLectura(args protocolo.ConsultaUltimaLectura, resp *protocolo.Lectura) error
func (t *Telemetria) ObtenerUltimaLectura(args protocolo.ConsultaUltimaLectura, resp *protocolo.Lectura) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	lectura, existe := t.lecturas[args.SensorID]

	if !existe {
		return fmt.Errorf("no se encontraron registros de telemetria para el sensor solicitado: '%s'", args.SensorID)
	}

	*resp = lectura
	log.Printf("[RPC-Server] Consulta procesada -> Sensor: %s recuperado con exito.\n", args.SensorID)

	return nil
}
