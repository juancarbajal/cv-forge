# Casos de Prueba Funcionales - CV Checker (V1)
## Plan y Matriz de Pruebas de Verificación de Requisitos Funcionales

---

## 1. Introducción y Alcance

Este documento define la suite de **Casos de Prueba (Test Cases)** para verificar de manera exhaustiva el cumplimiento de los **Requisitos Funcionales (RF-001 a RF-036)** especificados en [requirements.md](file:///home/jcarbajal/Projects/cv-checker/docs/requirements.md).

### Convención de Identificación
- **Formato de ID:** `TC-RF-XXX-NN` (donde `XXX` es el número de requisito funcional y `NN` es el número correlativo de la prueba).
- **Tipos de prueba:** Funcional (Positiva), Borde (Edge case), Negativa (Validación/Error).

---

## 2. Matriz de Trazabilidad RTM (Requirements Traceability Matrix)

| Módulo Funcional | Requisitos Cubiertos | Casos de Prueba Asociados |
| :--- | :--- | :--- |
| **Gestión y Almacenamiento de CVs** | RF-001, RF-002, RF-003, RF-004 | `TC-RF-001-01` a `TC-RF-004-02` |
| **Gestión de Postulaciones** | RF-005, RF-006, RF-007, RF-008 | `TC-RF-005-01` a `TC-RF-008-01` |
| **Análisis de Oferta Laboral** | RF-009, RF-010, RF-011 | `TC-RF-009-01` a `TC-RF-011-01` |
| **Matching Semántico y Evidencia** | RF-012, RF-013, RF-014 | `TC-RF-012-01` a `TC-RF-014-02` |
| **Score ATS Explicable** | RF-015, RF-016, RF-017 | `TC-RF-015-01` a `TC-RF-017-02` |
| **Recomendaciones y Anti-Invención** | RF-018, RF-019 | `TC-RF-018-01` a `TC-RF-019-02` |
| **Generación de CV y Traducción** | RF-020, RF-021, RF-022 | `TC-RF-020-01` a `TC-RF-022-02` |
| **Diff y Trazabilidad** | RF-023, RF-024 | `TC-RF-023-01` a `TC-RF-024-01` |
| **Gestión y Versionado de Prompts** | RF-025, RF-026, RF-027, RF-028, RF-029 | `TC-RF-025-01` a `TC-RF-029-02` |
| **Integración con IA y Proveedores** | RF-030, RF-031, RF-032, RF-033 | `TC-RF-030-01` a `TC-RF-033-01` |
| **Configuración y Cifrado de Claves** | RF-034, RF-035 | `TC-RF-034-01` a `TC-RF-035-02` |
| **Auditoría y Registro de Actividad** | RF-036 | `TC-RF-036-01` |

---

## 3. Especificación Detallada de Casos de Prueba

### Módulo 1: Gestión y Almacenamiento de CVs (RF-001 a RF-004)

#### `TC-RF-001-01`: Creación de CV manual
- **Requisito:** RF-001, RF-002
- **Precondiciones:** Aplicación abierta en la pestaña "Gestión de CVs".
- **Pasos:**
  1. Presionar el botón "Crear Nuevo CV".
  2. Ingresar Nombre: `"CV Desarrollador Go"`, Descripción: `"Perfil Backend"`, Idioma: `"Español (es)"`.
  3. Ingresar texto Markdown en el editor: `"# Juan Pérez\nExperiencia en Go y PostgreSQL."`.
  4. Presionar "Guardar".
- **Resultado Esperado:** 
  - El CV aparece en la lista izquierda.
  - Se genera automáticamente la versión `v1` marcada como favorita (`★`).
  - En la base de datos SQLite se crea el registro con ID único en la tabla `cvs`.
  - En el disco se crea el archivo físico `data/cvs/{cv_id}_v1.md` con el contenido exacto.

#### `TC-RF-001-02`: Importación de CV desde archivo Markdown externo
- **Requisito:** RF-001, RF-002
- **Precondiciones:** Archivo `mi_cv.md` existente en el sistema de archivos del usuario.
- **Pasos:**
  1. Presionar "Importar Markdown".
  2. Seleccionar el archivo `mi_cv.md`.
  3. Confirmar la carga.
- **Resultado Esperado:**
  - El contenido del archivo se lee y se almacena en `data/cvs/`.
  - Los metadatos se guardan en SQLite y el nuevo CV se visualiza en la interfaz.

#### `TC-RF-001-03`: Búsqueda y filtrado de CVs
- **Requisito:** RF-001
- **Precondiciones:** Existen al menos 3 CVs registrados ("CV Go Backend", "CV Frontend React", "CV DevOps").
- **Pasos:**
  1. Escribir `"React"` en la barra de búsqueda de CVs.
- **Resultado Esperado:**
  - La lista filtra instantáneamente mostrando únicamente el registro "CV Frontend React".
  - Al limpiar la barra de búsqueda, se restablece la lista completa.

#### `TC-RF-001-04`: Duplicar un CV existente
- **Requisito:** RF-001
- **Precondiciones:** CV "CV Go Backend" seleccionado con 2 versiones existentes.
- **Pasos:**
  1. Presionar el botón "Duplicar".
- **Resultado Esperado:**
  - Se crea un nuevo CV con título `"CV Go Backend (Copia)"`.
  - El contenido inicial del duplicado corresponde a la versión favorita del original.
  - El nuevo CV inicia con su propio historial independiente (`v1`).

#### `TC-RF-003-01`: Creación de nueva versión manual
- **Requisito:** RF-003
- **Precondiciones:** Un CV seleccionado en el editor.
- **Pasos:**
  1. Modificar el texto Markdown agregando una nueva experiencia.
  2. Ingresar en el campo motivo: `"Se agregó proyecto bancario"`.
  3. Presionar "Guardar como Nueva Versión".
- **Resultado Esperado:**
  - Se crea la versión `v2`.
  - En la lista de versiones se muestra `v2` con fecha, motivo y proveedor "Manual".
  - En disco se preserva tanto `data/cvs/{cv_id}_v1.md` como `data/cvs/{cv_id}_v2.md`.

#### `TC-RF-004-01`: Exclusividad de versión favorita (Regla 1 Favorita por CV)
- **Requisito:** RF-004
- **Precondiciones:** Un CV con versiones `v1` (actual favorita `★`), `v2` y `v3`.
- **Pasos:**
  1. Seleccionar la versión `v3`.
  2. Presionar el botón "Marcar Favorita (★)".
- **Resultado Esperado:**
  - La versión `v3` queda marcada con `★ Favorita`.
  - La versión `v1` pierde el estado de favorita.
  - En la tabla `cv_versions` únicamente `v3` tiene `is_favorite = 1`.
  - En la tabla `cvs`, `favorite_version_id` se actualiza al ID de `v3`.

---

### Módulo 2: Gestión de Postulaciones (RF-005 a RF-008)

#### `TC-RF-005-01`: Crear postulación laboral
- **Requisito:** RF-005
- **Precondiciones:** Existe al menos un CV registrado en el sistema.
- **Pasos:**
  1. Ir a la pestaña "Postulaciones & ATS".
  2. Presionar "Nueva Postulación".
  3. Completar campos:
     - Empresa: `"MercadoLibre"`
     - Cargo: `"Software Engineer Backend"`
     - URL: `"https://careers.mercadolibre.com/job/123"`
     - CV Base: Seleccionar `"CV Desarrollador Go"`
     - Idioma Objetivo: `"Español (es)"`
     - Anuncio: Pegar texto descriptivo de la vacante.
  4. Presionar "Crear".
- **Resultado Esperado:**
  - Se registra la postulación con estado `"Borrador"` y Score ATS `"0.0"`.
  - Aparece en el listado izquierdo ordenado cronológicamente.

#### `TC-RF-006-01`: Listar postulaciones con información resumen
- **Requisito:** RF-006
- **Precondiciones:** Existen múltiples postulaciones creadas.
- **Pasos:**
  1. Visualizar la lista de postulaciones.
- **Resultado Esperado:**
  - Cada elemento del listado muestra claramente: Empresa, Cargo, Estado actual y ATS Score.

#### `TC-RF-007-01`: Modificar datos de postulación existente
- **Requisito:** RF-007
- **Precondiciones:** Postulación seleccionada.
- **Pasos:**
  1. Cambiar el texto del anuncio laboral agregando nuevos requisitos.
  2. Cambiar el idioma objetivo a `"Inglés (en)"`.
  3. Presionar "Actualizar Datos de Postulación".
- **Resultado Esperado:**
  - Los datos se guardan en SQLite con timestamp `updated_at` actualizado.

#### `TC-RF-008-01`: Persistencia del historial completo de la postulación
- **Requisito:** RF-008
- **Precondiciones:** Postulación con análisis previo y optimización ejecutada.
- **Pasos:**
  1. Cerrar y volver a abrir la aplicación.
  2. Seleccionar la postulación analizada previamente.
- **Resultado Esperado:**
  - Se cargan y muestran intactos: los requisitos detectados, los resultados de matching con sus evidencias, el score ATS histórico desglosado y la versión generada vinculada.

---

### Módulo 3: Análisis de Oferta Laboral (RF-009 a RF-011)

#### `TC-RF-009-01`: Detección automática de idioma en la oferta
- **Requisito:** RF-009
- **Precondiciones:** Proveedor de IA configurado y postulación creada.
- **Pruebas por casos:**
  - **Caso A (Español):** Anuncio en español -> Detecta `"es"`.
  - **Caso B (Inglés):** Anuncio: `"We are looking for a Senior Go Engineer..."` -> Detecta `"en"`.
  - **Caso C (Portugués):** Anuncio: `"Procuramos desenvolvedor backend com experiência..."` -> Detecta `"pt"`.
- **Resultado Esperado:**
  - El campo `detected_language` se establece de forma precisa según el texto.

#### `TC-RF-010-01`: Extracción y clasificación de requisitos por prioridad
- **Requisito:** RF-010, RF-011
- **Precondiciones:** Anuncio laboral que especifique:
  - *"Obligatorio: 5 años con Go y Docker."*
  - *"Deseable: AWS y Kubernetes."*
  - *"Opcional: Experiencia en startups."*
- **Pasos:**
  1. Ejecutar "1. Analizar Oferta (Extraer Requisitos)".
- **Resultado Esperado:**
  - Se extraen los requisitos en una lista estructurada.
  - Cada requisito contiene: `nombre`, `categoria` (`hard_skill`, `experience`, `keywords`, `education`, `language`) y `prioridad` (`obligatorio`, `deseable`, `opcional`).

---

### Módulo 4: Matching CV vs Oferta y Evidencia (RF-012 a RF-014)

#### `TC-RF-012-01`: Clasificación tri-estado de requisitos
- **Requisito:** RF-012
- **Precondiciones:** Oferta con requisitos: `"Go"`, `"Docker"`, `"Ruby"`.
- **Datos de prueba (CV):** CV contiene mención de Go, mención de contenedores Linux (sin citar Docker explícitamente), y ninguna mención de Ruby.
- **Pasos:**
  1. Ejecutar "2. Evaluar Matching & Calcular ATS".
- **Resultado Esperado:**
  - Go -> Clasificado como `"Cumple"`.
  - Docker -> Clasificado como `"Cumple parcialmente"`.
  - Ruby -> Clasificado como `"No evidenciado"`.

#### `TC-RF-013-01`: Verificación de evidencia, nivel de confianza y explicación
- **Requisito:** RF-013
- **Pasos:**
  1. Inspeccionar los resultados de matching generados en `TC-RF-012-01`.
- **Resultado Esperado:**
  - Cada ítem contiene:
    - `evidencia`: Texto o referencia concreta extraída del CV (o vacío si no fue evidenciado).
    - `confianza`: Valor decimal entre `0.0` y `1.0` (ej: 0.95 para Go, 0.84 para Docker).
    - `explicacion`: Argumentación objetiva del resultado.

#### `TC-RF-014-01`: Evaluación flexible de sinónimos y equivalencias tecnológicas
- **Requisito:** RF-014
- **Casos de equivalencia a probar:**
  - Requisito Oferta: `"Golang"` <-> CV cita: `"Go"`.
  - Requisito Oferta: `"PostgreSQL"` <-> CV cita: `"Postgres"`.
  - Requisito Oferta: `"Kubernetes"` <-> CV cita: `"K8s"`.
  - Requisito Oferta: `"SQL Server"` <-> CV cita: `"MS SQL"`.
- **Resultado Esperado:**
  - La IA reconoce la equivalencia tecnológica y marca `"Cumple"` con alta confianza en lugar de marcar `"No evidenciado"`.

---

### Módulo 5: ATS Score Explicable (RF-015 a RF-017)

#### `TC-RF-015-01`: Rango y formato de puntuación ATS Total
- **Requisito:** RF-015
- **Pasos:**
  1. Ejecutar cálculo ATS sobre diversas combinaciones de CVs y ofertas.
- **Resultado Esperado:**
  - El puntaje total resultante está estrictamente comprendido en el rango continuo de `0.0` a `100.0`.

#### `TC-RF-016-01`: Desglose explicable por las 5 categorías
- **Requisito:** RF-016
- **Pasos:**
  1. Revisar la tarjeta de resultados ATS tras la evaluación.
- **Resultado Esperado:**
  - La UI y el registro muestran explícitamente la puntuación individual para:
    1. Hard Skills
    2. Experiencia
    3. Keywords ATS
    4. Educación
    5. Idioma

#### `TC-RF-017-01`: Cumplimiento estricto de ponderaciones fijas ATS
- **Requisito:** RF-017
- **Fórmula de verificación:**
  $$\text{Total} = (0.35 \times \text{Hard}) + (0.25 \times \text{Exp}) + (0.20 \times \text{KW}) + (0.10 \times \text{Edu}) + (0.10 \times \text{Lang})$$
- **Datos de prueba controlados:**
  - Hard Skills: 80.0
  - Experiencia: 100.0
  - Keywords: 70.0
  - Educación: 90.0
  - Idioma: 50.0
  - *Cálculo esperado:* $(0.35 \times 80) + (0.25 \times 100) + (0.20 \times 70) + (0.10 \times 90) + (0.10 \times 50) = 28 + 25 + 14 + 9 + 5 = \mathbf{81.0}$
- **Resultado Esperado:**
  - El total calculado por el motor coincide con `81.0` sin discrepancias de redondeo arbitrario.

---

### Módulo 6: Recomendaciones y Restricción Anti-Invención (RF-018, RF-019)

#### `TC-RF-018-01`: Generación estructurada de recomendaciones
- **Requisito:** RF-018
- **Pasos:**
  1. Verificar la sección de recomendaciones tras el análisis ATS.
- **Resultado Esperado:**
  - Contiene listas no vacías para:
    - Mejoras recomendadas.
    - Riesgos detectados frente a la vacante.
    - Oportunidades.
    - Requisitos faltantes.
    - Keywords ATS estratégicas.

#### `TC-RF-019-01`: Detección y bloqueo de invenciones no sustentadas (Anti-Hallucination)
- **Requisito:** RF-019
- **Escenario:** Un CV que NO menciona idiomas extranjeros ni certificaciones Cloud.
- **Pasos:**
  1. Simular respuesta de optimización donde se intente inyectar: `"Certificación AWS Solutions Architect"` o `"Inglés C2 bilingüe"` sin evidencia previa.
  2. Ejecutar el validador de dominio `AntiInventionValidator`.
- **Resultado Esperado:**
  - El validador detecta la falta de sustento en el CV origen.
  - Se genera una alerta o bloqueo impidiendo guardar la invención sin justificación trazable.

---

### Módulo 7: Generación de CV y Traducción (RF-020 a RF-022)

#### `TC-RF-020-01`: Generación de CV optimizado en Markdown
- **Requisito:** RF-020
- **Pasos:**
  1. Presionar "3. Generar CV Optimizado con IA".
- **Resultado Esperado:**
  - Se genera un documento Markdown completo, coherente y enriquecido que reorganiza y resalta las fortalezas del candidato alineadas a la oferta.
  - Se registra automáticamente como una nueva versión en el CV base (`v(n+1)`).

#### `TC-RF-021-01`: Selección de idioma de salida
- **Requisito:** RF-021
- **Pasos:**
  1. Postulación con CV origen en Español y oferta en Inglés.
  2. Seleccionar Idioma Objetivo: `"Inglés (en)"`.
  3. Ejecutar optimización.
- **Resultado Esperado:**
  - El CV generado se encuentra completamente redactado en Inglés.

#### `TC-RF-022-01`: Traducción técnica preservando contexto profesional
- **Requisito:** RF-022
- **Pasos:**
  1. Verificar la traducción de secciones clave: Resumen profesional, Experiencia laboral, Educación, Habilidades y Certificaciones.
- **Resultado Esperado:**
  - Los términos técnicos estándar (ej: "Backend Developer", "Framework", "Microservices", "CI/CD") se conservan adecuadamente sin traducciones literales deformadas.

---

### Módulo 8: Diff Estilo Git y Trazabilidad (RF-023, RF-024)

#### `TC-RF-023-01`: Generación de Diff Unificado estilo Git
- **Requisito:** RF-023
- **Pasos:**
  1. Visualizar la pestaña "Optimización & Diff" tras generar la nueva versión.
- **Resultado Esperado:**
  - El visor muestra las diferencias unificadas:
    - Líneas eliminadas o reemplazadas prefijadas con `-`.
    - Líneas añadidas o mejoradas prefijadas con `+`.
    - Líneas sin cambio prefijadas con espacio de contexto.

#### `TC-RF-024-01`: Trazabilidad detallada por cambio realizado
- **Requisito:** RF-024
- **Pasos:**
  1. Revisar la lista de cambios en el panel lateral de trazabilidad.
- **Resultado Esperado:**
  - Cada tarjeta de cambio contiene:
    1. Resumen del cambio.
    2. Motivo del cambio.
    3. Requisito de la oferta relacionado.
    4. Impacto ATS estimado (ej: `+3.5 pts`).
    5. Evidencia del CV original utilizada para sustentar el cambio.

---

### Módulo 9: Gestión y Versionado de Prompts (RF-025 a RF-029)

#### `TC-RF-025-01`: Presencia de prompts iniciales por tipo (Seed Defaults)
- **Requisito:** RF-025
- **Precondiciones:** Primera ejecución con base de datos limpia.
- **Pasos:**
  1. Abrir la pestaña "Gestión de Prompts".
  2. Desplegar el selector de tipos de prompts.
- **Resultado Esperado:**
  - Existen prompts preconfigurados para: `JOB_ANALYSIS`, `CV_MATCHING`, `ATS_SCORING`, `CV_OPTIMIZATION` y `TRANSLATION`.

#### `TC-RF-026-01`: Creación de nueva versión de prompt
- **Requisito:** RF-026, RF-027
- **Pasos:**
  1. Seleccionar tipo `JOB_ANALYSIS`.
  2. Modificar el texto del prompt en el editor.
  3. Ingresar nombre: `"Ajuste para roles DevOps"`.
  4. Presionar "Guardar como Nueva Versión (v+1)".
- **Resultado Esperado:**
  - Se genera la versión `v2`.
  - Se preserva el registro de la versión `v1`.

#### `TC-RF-028-01`: Exclusividad de prompt activo por tipo
- **Requisito:** RF-028
- **Pasos:**
  1. Crear la versión `v2` con la casilla "Establecer inmediatamente como versión activa" marcada.
- **Resultado Esperado:**
  - `v2` queda con la etiqueta `★ ACTIVA`.
  - `v1` pasa a estado inactivo. En SQLite solo una fila tiene `is_active = 1` para ese `prompt_type`.

#### `TC-RF-029-01`: Rollback y restauración de versión anterior de prompt
- **Requisito:** RF-029
- **Pasos:**
  1. Seleccionar la versión anterior `v1`.
  2. Presionar "Restaurar Versión Seleccionada (Rollback)".
- **Resultado Esperado:**
  - `v1` vuelve a ser la versión activa para futuras ejecuciones de IA.

#### `TC-RF-029-02`: Comparación Diff entre versiones de prompts
- **Requisito:** RF-029
- **Pasos:**
  1. Seleccionar una versión del prompt y presionar "Comparar con Versión Activa (Diff)".
- **Resultado Esperado:**
  - El visor inferior muestra el diff estilo Git reflejando exactamente qué instrucciones del prompt fueron añadidas o modificadas.

---

### Módulo 10: Integración con IA y Configuración Segura (RF-030 a RF-035)

#### `TC-RF-030-01`: Selección de proveedor de IA soportado
- **Requisito:** RF-030
- **Pasos:**
  1. Abrir la pestaña "Configuración IA".
  2. Desplegar el selector de proveedores.
- **Resultado Esperado:**
  - Permite seleccionar: `gemini`, `groq`, `openai`, `ollama` y `mock`.

#### `TC-RF-031-01`: Cambio dinámico de proveedor en caliente
- **Requisito:** RF-031
- **Pasos:**
  1. Estar utilizando `gemini` como proveedor activo.
  2. Cambiar el proveedor activo a `ollama` (o `mock`) en la interfaz.
  3. Ejecutar un nuevo análisis en una postulación sin reiniciar la app.
- **Resultado Esperado:**
  - El análisis se procesa inmediatamente con el nuevo proveedor seleccionado sin necesidad de recompilar ni reiniciar.

#### `TC-RF-032-01`: Modelos configurables por proveedor
- **Requisito:** RF-032
- **Pasos:**
  1. Seleccionar `gemini` y especificar modelo `"gemini-2.0-flash"`.
  2. Seleccionar `groq` y especificar `"llama-3.3-70b-versatile"`.
  3. Seleccionar `ollama` y especificar `"qwen2.5:latest"`.
- **Resultado Esperado:**
  - La aplicación almacena el modelo indicado y lo envía en el payload correspondiente.

#### `TC-RF-034-01`: Gestión de API Keys
- **Requisito:** RF-034
- **Pasos:**
  1. Ingresar una API key para OpenAI.
  2. Guardar la configuración.
- **Resultado Esperado:**
  - La configuración queda registrada y disponible para llamadas al servicio.

#### `TC-RF-035-01`: Cifrado autenticado local con AES-256 GCM
- **Requisito:** RF-035
- **Pasos:**
  1. Guardar una API key secreta (ej: `"sk-proj-super-secret-12345"`).
  2. Inspeccionar directamente el archivo SQLite `data/cvchecker.db` con una consulta SQL:
     `SELECT encrypted_api_key FROM provider_configs WHERE provider_name = 'openai';`
- **Resultado Esperado:**
  - El campo contiene un blob binario cifrado no legible.
  - En ningún lugar de la base de datos ni archivos temporales existe la clave en texto plano.
  - Al realizar una petición, la aplicación descifra en memoria el token exitosamente usando la llave derivada localmente.

---

### Módulo 11: Auditoría y Registro de Actividad (RF-036)

#### `TC-RF-036-01`: Registro inmutable de acciones en Auditoría
- **Requisito:** RF-036
- **Pasos:**
  1. Realizar una optimización de CV y un cambio de configuración de proveedor.
  2. Abrir la pestaña "Auditoría".
- **Resultado Esperado:**
  - Cada acción aparece listada en orden cronológico inverso con:
    - Fecha y hora UTC precisa.
    - Nombre de la acción (`OPTIMIZE_CV`, `UPDATE_CONFIG`, etc.).
    - Proveedor de IA y modelo utilizado.
    - Estado de resultado (`SUCCESS` o `FAILED`).
    - Metadatos JSON descriptivos de la operación.

---

## 4. Guía de Ejecución de Pruebas Automatizadas

Para validar los casos de prueba automatizables en el código, ejecutar los siguientes comandos desde la raíz del proyecto:

```bash
# 1. Pruebas de Dominio (ATS Score determinista, Pesos fijos, Anti-Invención)
go test -v ./internal/domain/...

# 2. Pruebas de Cifrado AES-256 GCM y Seguridad
go test -v ./internal/infrastructure/security/...

# 3. Pruebas de Persistencia SQLite y Almacenamiento Markdown
go test -v ./internal/infrastructure/persistence/...

# 4. Pruebas de Motor Diff estilo Git
go test -v ./internal/infrastructure/diff/...

# 5. Pruebas de Factoría y Proveedores de IA
go test -v ./internal/infrastructure/ai/...

# 6. Pruebas de Flujo Completo de Casos de Uso (Application Services)
go test -v ./internal/app/...

# 7. Ejecución de Toda la Suite
go test -v ./...
```
