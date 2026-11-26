# Documento de Requisitos Funcionales y No Funcionales

## Sistema de Optimización Inteligente de CV para Postulaciones Laborales (V1)

---

# 1. Introducción

## 1.1 Objetivo

Desarrollar una aplicación de escritorio que permita gestionar múltiples CVs en formato Markdown y optimizarlos automáticamente para ofertas laborales específicas utilizando Inteligencia Artificial.

La solución deberá:

- Analizar ofertas laborales.
- Comparar requisitos contra un CV seleccionado.
- Calcular un score ATS explicable.
- Detectar coincidencias, faltantes y coincidencias parciales.
- Sugerir mejoras trazables.
- Generar nuevas versiones optimizadas del CV.
- Gestionar historial de postulaciones.
- Mantener versionado completo de CVs y prompts.
- Permitir cambiar fácilmente de proveedor de IA.

---

# 2. Alcance

## Incluido en la V1

- Gestión de CVs.
- Gestión de versiones.
- Gestión de postulaciones.
- Integración con IA.
- Traducción automática.
- Análisis ATS.
- Generación automática de CV optimizado.
- Historial de análisis.
- Diff estilo Git.
- Soporte para múltiples proveedores IA.
- Versionado de prompts.

## No incluido en la V1

- Exportación PDF.
- Sincronización en nube.
- Multiusuario.
- Integración con portales de empleo.
- OCR.
- Integración con LinkedIn.

---

# 3. Requisitos Funcionales

## RF-001 Gestión de CVs

El sistema deberá permitir:

- Crear un CV.
- Importar un CV Markdown.
- Editar un CV.
- Eliminar un CV.
- Duplicar un CV.
- Marcar un CV como favorito.
- Buscar CVs.

### Campos

| Campo | Descripción |
|---------|---------|
| ID | Identificador único |
| Nombre | Nombre del CV |
| Descripción | Descripción opcional |
| Idioma | Idioma principal |
| Fecha de creación | Timestamp |
| Fecha de actualización | Timestamp |
| Favorito | Boolean |

---

## RF-002 Almacenamiento de CVs

El sistema deberá almacenar:

### Metadatos

En SQLite.

### Contenido

Como archivo físico:

```text
data/cvs/
```

Formato:

```text
*.md
```

---

## RF-003 Versionado*de CV

El sistema deberá crear una*nueva versión cuando:

- El usuari* lo solicite.
- Una IA genere un C* optimizado.

Cada versión deberá *onservar:

```text
Número de versi*n
Fecha
CV origen
CV generado
Moti*ación del cambio
Proveedor IA util*zado
Modelo utilizado
```

---

##*RF-004 Gestión de Favoritos

El us*ario podrá marcar únicamente una v*rsión favorita por CV.

Ejemplo:

*``text
CV Backend

v1
v2
v3 ★
v4
`*`

---

# Gestión de Postulaciones*
## RF-005 Crear Postulación

El u*uario deberá poder:

- Crear una p*stulación.
- Asignar nombre.
- Aso*iar empresa.
- Asociar cargo.
- As*ciar anuncio laboral.
- Selecciona* un CV.

*## Datos mínimos

```text
Empresa
*argo
URL (opcional)
Anuncio labora*
CV seleccionado
Idioma objetivo
`*`

---

## RF-006 Listar Postulaci*nes

El sistema deberá mostrar:

`*`text
Empresa
Cargo
Estado
ATS Sco*e
Fecha creación
Fecha actualizaci*n
```

---

## RF-007 Editar Post*lación

El usuario podrá:

- Modif*car anuncio.
- Cambiar CV.
- Cambi*r idioma objetivo.
- Volver a ejec*tar análisis.

---

*# RF-008 Historial de Postulacione*

El sistema deberá conservar:

- *nálisis realizados.
- Versiones ge*eradas.
- Resultados ATS histórico*.
- Resultados de matching.
- Reco*endaciones generadas.

---

# Anál*sis de Oferta

## RF-009 Detección*de Idioma

El sistema deberá detec*ar automáticamente:

- Español.
- *nglés.
- Portugu*s.

---

## RF-010 Extracción de R*quisitos

La IA deberá extraer:

#*# Obligatorios

Ejemplo:

```text
*o
Docker
5 años de experiencia
```*
### Deseables

Ejemplo:

```text
*WS
Kubernetes
```

### Opcionales
*Ejemplo:

```text
Experiencia en s*artups
```

---

## RF-011 Clasifi*ación de Requisitos

Cada requisit* deberá contener:

```json
{
  "no*bre": "Docker",
  "categoria": "ha*d_skill",
  "prioridad": "obligato*io"
}
```

---

# Matching CV

## *F-012 Comparación CV vs Oferta

El*sistema deberá clasificar cada req*isito como:

```text
Cumple
Cumple*parcialmente
No evidenciado
```

-*-

## RF-013 Evidencia

Para cada *equisito identificado deberá exist*r:

```text
Evidencia encontrada
N*vel de confianza
Explicación
```

*jemplo:

```json
{
  "requisito": *Docker",
  "estado": "Parcial",
  *confianza": 0.84,
  "evidencia": "*xperiencia con contenedores Linux"*}
```

---

## RF-014 Evaluación F*exible

El sistema deberá reconoce*:

- Sinónimos.
- Variaciones tecn*lógicas.
- Equivalencias contextua*es.

Ejemplos:

```text
Go = Golan*
Postgres = PostgreSQL
K8s = Kuber*etes
MS SQL = SQL Server
```

---
*# ATS Score

## RF-015 Cálculo ATS*
El sistema deberá calcular:

```t*xt
ATS Total
```

Escala:

```text*0-100
```

---

## RF-016 Score Ex*licable

El ATS deberá mostrar:

`*`text
Hard Skills
Experiencia
Keyw*rds
Educación
Idioma
```

---

## *F-017 Pesos ATS

Los pesos serán f*jos:

| Categoría | Peso |
|------*-----|-------|
| Hard Skills | 35%*|
| Experiencia | 25% |
| Keywords*ATS | 20% |
| Educación | 10% |
| *dioma | 10% |

---

# Recomendacio*es

## RF-018 Generación de Recome*daciones

La IA deberá generar:

-*Mejoras recomendadas.
- Riesgos.
-*Oportunidades.
- Faltantes.
- Keyw*rds ATS relevantes.

---

## RF-01* Restricción Anti-Invención

La IA*NO podrá:

- Inventar experiencia *aboral.
- Inventar tecnologías.
- *nventar certificaciones.
- Inventa* idiomas.
- Inventar educación.
- *nventar proyectos.

Toda mejora de*erá derivarse de evidencia existen*e.

---

# Generación de CV

## RF*020 Generación Automática

La IA d*berá generar:

```text
CV optimiza*o
```

considerando:

```text
CV o*iginal
+
Oferta laboral
+
Idioma o*jetivo
```

---

## RF-021 Idioma *e Salida

Por defecto:

```text
Id*oma detectado en la oferta laboral*```

Opciones:

```text
Español
In*lés
Portugués
```

---

## RF-022 *raducción

La IA deberá traducir:
*- Resumen profesional.
- Experienc*a laboral.
- Educación.
- Habilida*es.
- Certificaciones.

Preservand* contexto profesional.

---

# Dif*

## RF-023 Generación de Diff

El*sistema deberá generar:

```diff
-*Experience developing backend serv*ces.
+ Experience developing backe*d services using Go and PostgreSQL*
```

---

## RF-024 Trazabilidad
*Cada cambio deberá indicar:

```te*t
Motivo
Requisito relacionado
Imp*cto ATS esperado
```

Ejemplo:

``*text
Cambio:
Se agregó PostgreSQL.*
Motivo:
Tecnología obligatoria de*ectada.

Impacto:
+4 ATS Score.

E*idencia:
El CV ya mencionaba bases*de datos relacionales.
```

---

#*Gestión de Prompts

## RF-025 Catá*ogo de Prompts

El sistema deberá *isponer de prompts separados para:*
```text
Análisis de oferta
Matchi*g
ATS
Traducción
Optimización de C*
Extracción de requisitos
Clasific*ción de requisitos
```

---

## RF*026 Versionado de Prompts

Cada pr*mpt deberá almacenar:

```text
ID
Nombre
Tipo
Versión
Fecha creación
Activo
Contenido
```

---

## RF-027 Historial de Prompts

El sistema deberá conservar:

```text
v1
v2
v3
v4
...
```

---

## RF-028 Prompt Activo

Solo una versión podrá estar activa por tipo de prompt.

Ejemplo:

```text
Optimización CV

v1
v2
v3 ★ Activa
```

---

## RF-029 Recuperación de Prompts

El usuario podrá:

- Consultar versiones anteriores.
- Restaurar una versión previa.
- Comparar cambios entre prompts.

---

# Integración con IA

## RF-030 Selección de Proveedor

El usuario podrá seleccionar:

```text
Gemini
Groq
OpenAI
Ollama
```

---

## RF-031 Cambio Dinámico

El cambio de proveedor no requerirá:

- Recompilar.
- Modificar código.
- Migrar datos.

---

## RF-032 Modelos Configurables

Ejemplos:

### Gemini

```text
Gemini Flash
Gemini Pro
```

### Groq

```text
Llama
Mixtral
DeepSeek
```

### Ollama

```text
Qwen
Llama
Mistral
Phi
DeepSeek
```

---

## RF-033 Arquitectura de Proveedores

El sistema utilizará una interfaz desacoplada:

```go
type AIProvider interface {
    AnalyzeJob(...)
    MatchCV(...)
    GenerateOptimizedCV(...)
    Translate(...)
}
```

---

# Configuración

## RF-034 Claves API

La aplicación deberá permitir:

- Agregar claves.
- Editar claves.
- Eliminar claves.
- Activar proveedor.

---

## RF-035 Almacenamiento Seguro

Las claves deberán almacenarse:

- Cifradas.
- Localmente.
- En SQLite.

---

# Auditoría

## RF-036 Registro de Actividad

El sistema deberá registrar:

```text
Fecha
Hora
Acción
Proveedor IA
Modelo utilizado
Resultado
```

---

# Experiencia de Usuario y Accesibilidad

## RF-037 Etiquetado y Claridad en Formularios

El sistema deberá asegurar que:

- Todo campo de entrada (`Entry`), selector (`Select`) y área de texto (`MultiLineEntry`) cuente con un `label` visible y descriptivo que indique inequívocamente al usuario el dato solicitado.
- Los campos contengan textos de ayuda (`placeholder`) con ejemplos concretos de formato o contenido esperado.
- Ningún control interactivo quede sin identificación textual visible inmediata en la interfaz.

---

# 4. Requisitos No Funcionales

## RNF-001 Plataforma

Aplicación de escritorio desarrollada con:

```text
Go
Fyne
```

---

## RNF-002 Base de Datos

Motor:

```text
SQLite
```

---

## RNF-003 Sistema Operativo

Compatibilidad:

```text
Windows
Linux
macOS
```

---

## RNF-004 Arquitectura

La solución deberá seguir una arquitectura por capas:

```text
UI
Application
Domain
Providers
Repositories
Persistence
```

---

## RNF-005 Desacoplamiento IA

La lógica de negocio no dependerá de un proveedor específico.

Deberá soportar:

```text
Gemini
Groq
OpenAI
Ollama
```

mediante Strategy + Factory Pattern.

---

## RNF-006 Almacenamiento Local

Toda la información deberá permanecer local.

No se utilizarán servicios cloud para persistencia.

---

## RNF-007 Seguridad

Las claves API deberán cifrarse usando:

```text
AES-256 GCM
```

---

## RNF-008 Rendimiento

Tiempo objetivo para una ejecución completa:

```text
Menor a 60 segundos
```

para:

```text
CV estándar
Oferta estándar
```

dependiendo del proveedor IA.

---

## RNF-009 Resiliencia

Errores de IA no deberán afectar:

```text
CVs
Versiones
Postulaciones
Base de datos
Prompts
```

---

## RNF-010 Mantenibilidad

Los prompts deberán poder actualizarse sin recompilar la aplicación.

---

## RNF-011 Auditabilidad

Toda acción relevante deberá quedar registrada.

---

## RNF-012 Escalabilidad

La arquitectura deberá permitir futuras extensiones:

```text
PDF
DOCX
Nuevos idiomas
Nuevos motores ATS
Nuevos proveedores IA
Nuevos workflows
```

---

## RNF-013 Testabilidad

Los servicios deberán diseñarse mediante interfaces para facilitar:

```text
Unit Testing
Integration Testing
Mock Testing
```

---

## RNF-014 Portabilidad

La aplicación deberá compilarse desde una única base de código para:

```text
Windows
Linux
macOS
```

---

# 5. Criterios de Éxito de la V1

La V1 será considerada exitosa cuando permita:

1. Gestionar múltiples CVs Markdown.
2. Versionar CVs.
3. Marcar versiones favoritas.
4. Gestionar postulaciones.
5. Analizar ofertas laborales.
6. Clasificar requisitos en Obligatorios, Deseables y Opcionales.
7. Detectar idioma automáticamente.
8. Traducir entre Español, Inglés y Portugués.
9. Calcular ATS Score explicable.
10. Generar recomendaciones trazables.
11. Generar CVs optimizados.
12. Generar diffs estilo Git.
13. Mantener historial de análisis.
14. Mantener historial de postulaciones.
15. Mantener versionado de prompts.
16. Permitir cambiar entre Gemini, Groq, OpenAI y Ollama sin modificar la lógica de negocio.
17. Mantener toda la información almacenada localmente.
