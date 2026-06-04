package mqtt

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"sd-iot/pkg/nodo"
	"sd-iot/pkg/sensor"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// Cliente encapsula la conexión MQTT del nodo.
type Cliente struct {
	config   nodo.Configuracion
	interno  mqtt.Client
	opciones *mqtt.ClientOptions
}

// TODO 1: NuevoCliente crea la configuración inicial del cliente MQTT.
func NuevoCliente(config nodo.Configuracion) (*Cliente, error) {
	// 1a. Construir el tópico del testamento: nodo/{id}/estado
	topicoEstado := fmt.Sprintf("nodo/%s/estado", config.ID)

	// 1b. Configurar el mensaje del testamento como: {"estado":"offline"} con QoS 1 y retained=true.
	payloadTestamento := `{"estado":"offline"}`

	// 1c. Configurar ClienteID único, timeout de conexión, reconexión automática.

	opts := mqtt.NewClientOptions().
		AddBroker(config.BrokerURL).
		SetClientID(config.ID). // ClientID único
		SetConnectTimeout(5 * time.Second).
		SetAutoReconnect(true). // Reconexión automática
		SetCleanSession(false)

	opts.SetWill(topicoEstado, payloadTestamento, 1, true)

	return &Cliente{
		config:   config,
		opciones: opts,
	}, nil
}

// TODO 2: Conectar establece la sesión con el broker.
func (c *Cliente) Conectar() error {

	c.interno = mqtt.NewClient(c.opciones)

	token := c.interno.Connect()
	if token.WaitTimeout(10*time.Second) && token.Error() != nil {
		return fmt.Errorf("error al conectar a MQTT: %w", token.Error())
	}

	log.Println("[MQTT] Conexion establecida de forma segura.")

	// Tras conectar, publicar el mensaje retenido {"estado":"online"} en nodo/{id}/estado
	topicoEstado := fmt.Sprintf("nodo/%s/estado", c.config.ID)
	payloadOnline := `{"estado":"online"}`

	tokenPub := c.interno.Publish(topicoEstado, 1, true, payloadOnline)
	tokenPub.WaitTimeout(5 * time.Second)

	return nil
}

// TODO 3: PublicarLecturas envía periódicamente las lecturas del sensor.
func (c *Cliente) PublicarLecturas(sim *sensor.Simulador, config nodo.Configuracion) {
	// 3a. Construir el tópico: campus/{edificio}/{aula}/sensor/temperatura
	topicoSensor := fmt.Sprintf("campus/%s/%s/sensor/temperatura", config.Edificio, config.Aula)

	type PayloadLectura struct {
		NodoID      string    `json:"nodo_id"`
		Temperatura float64   `json:"temperatura"`
		Unidad      string    `json:"unidad"`
		Timestamp   time.Time `json:"timestamp"`
	}

	// Ticker
	intervalo := config.IntervaloSegundos
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	log.Printf("[MQTT] Iniciando loop de telemetría cada %v\n", intervalo)

	for range ticker.C {
		if !c.interno.IsConnected() {
			log.Println("[MQTT-Warning] Intentando publicar pero el cliente está desconectado. Esperando reconexión...")
			continue
		}

		// 3b. Llamar sim.Leer()
		valorTemperatura := sim.Leer()

		lectura := PayloadLectura{
			NodoID:      config.ID,
			Temperatura: valorTemperatura,
			Unidad:      "C",
			Timestamp:   time.Now(),
		}

		// Serializar a JSON
		jsonData, err := json.Marshal(lectura)
		if err != nil {
			log.Printf("[MQTT-Error] Error al serializar lectura: %v\n", err)
			continue
		}

		// Publicar con QoS 1
		token := c.interno.Publish(topicoSensor, 1, false, jsonData)

		go func(t mqtt.Token, data []byte) {
			if t.WaitTimeout(5*time.Second) && t.Error() != nil {
				log.Printf("[MQTT-Error] Falla al enviar paquete de telemetría: %v\n", t.Error())
			} else {
				log.Printf("[MQTT-Pub] %s -> %s\n", topicoSensor, string(data))
			}
		}(token, jsonData)
	}
}

// TODO 4: SuscribirComandos se une al tópico de actuadores y procesa mensajes.
func (c *Cliente) SuscribirComandos(config nodo.Configuracion) error {
	// 4a. Tópico: campus/{edificio}/{aula}/actuador/cmd
	topicoCmd := fmt.Sprintf("campus/%s/%s/actuador/cmd", config.Edificio, config.Aula)

	// Estructura para parsear el comando entrante
	type ComandoActuador struct {
		Accion string `json:"accion"`
		Origen string `json:"origen"`
	}

	// 4b.
	callbackMessage := func(client mqtt.Client, msg mqtt.Message) {
		var cmd ComandoActuador
		if err := json.Unmarshal(msg.Payload(), &cmd); err != nil {
			log.Printf("[MQTT-Sub] Error decodificando comando JSON: %v. Payload: %s\n", err, string(msg.Payload()))
			return
		}

		log.Printf("[MQTT-Cmd Recibido] Ejecutando accion '%s' solicitada por '%s'\n", cmd.Accion, cmd.Origen)

		switch cmd.Accion {
		case "encender_alarma":
			fmt.Println("¡Alerta! Encendiendo sistema de alarma del aula...")
		case "apagar_alarma":
			fmt.Println(" Alarma desactivada.")
		default:
			fmt.Printf(" Comando desconocido o no soportado: %s\n", cmd.Accion)
		}
	}

	token := c.interno.Subscribe(topicoCmd, 1, callbackMessage)
	if token.WaitTimeout(5*time.Second) && token.Error() != nil {
		return fmt.Errorf("error al suscribirse a %s: %w", topicoCmd, token.Error())
	}

	log.Printf("[MQTT] Suscripto con éxito al canal de comandos: %s\n", topicoCmd)
	return nil
}

// TODO 5: Desconectar cierra limpiamente la sesión MQTT.
func (c *Cliente) Desconectar() {
	if c.interno != nil && c.interno.IsConnected() {
		// Sugerencia: Publicar estado offline de forma explícita y retenida antes de salir
		topicoEstado := fmt.Sprintf("nodo/%s/estado", c.config.ID)
		payloadOffline := `{"estado":"offline"}`

		log.Println("[MQTT] Notificando estado offline manual antes del cierre...")
		token := c.interno.Publish(topicoEstado, 1, true, payloadOffline)
		token.WaitTimeout(2 * time.Second)

		c.interno.Disconnect(250)
		log.Println("[MQTT] Cliente desconectado limpiamente.")
	}
}
