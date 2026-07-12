# Firmar y notarizar Andromeda para macOS — guía paso a paso

> Estado: **pendiente de ejecución** (requiere una identidad de Apple Developer de pago).
> El *wiring* ya está hecho en `.goreleaser.yaml` (bloque `notarize:`, activado por credenciales);
> esta guía explica qué falta, por qué, y las condiciones legales/comerciales.

Esta guía es informativa. No es asesoría legal ni fiscal. Los puntos de impuestos, registro de
empresa y propiedad intelectual dependen de **tu país de residencia/constitución**, que aquí no se
asume; donde eso importa, lo señalo explícitamente y conviene confirmarlo con un profesional local.

---

## 1. Qué es esto y por qué se requiere

macOS incluye **Gatekeeper**: al abrir por primera vez un binario descargado de internet, el
sistema verifica dos cosas antes de dejarlo ejecutar:

1. **Firma de código (code signing)** con un certificado **Developer ID Application** emitido por
   Apple. Prueba *quién* produjo el binario y que no se alteró desde que lo firmaste.
2. **Notarización (notarization)**: subes el binario al servicio de Apple, que lo escanea
   automáticamente en busca de malware y componentes maliciosos y devuelve un "ticket". Ese ticket
   se **engrapa (staple)** al artefacto.

Sin firma + notarización, en macOS moderno (Catalina 10.15 en adelante) el usuario ve el bloqueo
*"no se puede abrir porque Apple no puede comprobar que no contiene malware"* y tiene que ir a
Ajustes del Sistema → Privacidad y seguridad → "Abrir de todos modos", o quitar el atributo de
cuarentena por terminal. Es fricción y desconfianza para cada usuario nuevo.

**Importante:** notarizar **NO** es lo mismo que publicar en la Mac App Store. La notarización
habilita la **distribución directa** (tu web, GitHub Releases, Homebrew) con la experiencia
"simplemente funciona". La App Store es un canal aparte con revisión humana y comisión (ver §7).

---

## 2. Qué necesitas exactamente

| Requisito | Detalle |
|---|---|
| **Apple ID** con verificación en dos pasos (2FA) | Gratis. Es la cuenta base. |
| **Membresía Apple Developer Program** | **De pago** (ver §4). Sin esto no puedes emitir un certificado Developer ID. |
| **Certificado "Developer ID Application"** | Se genera desde tu cuenta de desarrollador; es el que firma binarios para distribución **fuera** de la App Store. |
| **Credencial para `notarytool`** | Una **API Key de App Store Connect** (recomendada) — un `.p8` + Key ID + Issuer ID — o un Apple ID + contraseña específica de app. |
| **Un Mac** para firmar | El firmado/notarizado corre en macOS (o en un runner macOS de CI). Cross-compilas los binarios en cualquier sitio, pero firmarlos requiere las herramientas de Apple. |

Andromeda ya trae el bloque `notarize:` en `.goreleaser.yaml` gated por `MACOS_SIGN_P12`, así que
la integración es principalmente **proveer esas credenciales**.

---

## 3. Pasos, en orden

1. **Inscríbete en el Apple Developer Program** (§4). Como **individuo** o como **organización**
   (empresa). La diferencia importa para el nombre que aparece como firmante y para §6/§8.
2. **Crea el certificado Developer ID Application:**
   - En `developer.apple.com` → Certificates, IDs & Profiles → Certificates → `+` → *Developer ID
     Application*, o con Xcode (Settings → Accounts → Manage Certificates).
   - Expórtalo como `.p12` (incluye la clave privada) protegido con contraseña. Ese `.p12` y su
     contraseña son los secretos `MACOS_SIGN_P12` / `MACOS_SIGN_PASSWORD` del pipeline.
3. **Crea una API Key de App Store Connect para notarizar:**
   - `appstoreconnect.apple.com` → Users and Access → Integrations/Keys → genera una key con rol
     *Developer*. Descarga el `.p8` (¡solo se descarga una vez!). Anota **Key ID** e **Issuer ID**.
   - Estos son los secretos `MACOS_NOTARY_KEY` (contenido del `.p8`), `MACOS_NOTARY_KEY_ID`,
     `MACOS_NOTARY_ISSUER_ID` del pipeline.
4. **Prueba local** (una vez), sobre un binario ya construido:
   ```bash
   # Firmar
   codesign --sign "Developer ID Application: TU NOMBRE (TEAMID)" --options runtime --timestamp ./andromeda
   # Empaquetar y notarizar
   ditto -c -k --keepParent ./andromeda ./andromeda.zip
   xcrun notarytool submit ./andromeda.zip --key AuthKey_XXXX.p8 --key-id KEYID --issuer ISSUERID --wait
   # Engrapar el ticket (para binarios sueltos se engrapa el contenedor .dmg/.pkg; para un CLI
   # suelto el ticket viaja "online", pero para .dmg/.pkg: xcrun stapler staple ./andromeda.dmg)
   ```
5. **Configura los secretos en CI** (GitHub Actions → Settings → Secrets) con los mismos nombres
   que espera `.goreleaser.yaml`. En cuanto `MACOS_SIGN_P12` exista, el bloque `notarize:` se
   **activa solo** (`enabled: '{{ isEnvSet "MACOS_SIGN_P12" }}'`) y el release firma+notariza.
6. **Verifica** en una Mac limpia:
   ```bash
   spctl -a -vvv -t install ./andromeda      # debe decir "accepted / Notarized Developer ID"
   codesign --verify --deep --strict --verbose=2 ./andromeda
   ```

> Nota de CI: firmar/notarizar necesita un **runner macOS**, que en GitHub Actions consume minutos
> a **10×** el ratio de Linux. En el plan gratuito los minutos de Actions son limitados; considera
> firmar solo en releases de versión (no en cada push) o hacerlo localmente en tu Mac.

---

## 4. Pagos, suscripción y condiciones

- **Costo:** el Apple Developer Program cuesta **99 USD al año** (renovación anual; el precio local
  puede variar por impuestos/*foreign exchange*). Es **suscripción**: si no renuevas, tus
  certificados dejan de emitir y no puedes notarizar builds nuevos (los ya notarizados siguen
  válidos).
- **Programas sin costo** existen para organizaciones sin ánimo de lucro/educación acreditadas
  (*Apple Developer Program fee waiver*), sujeto a aprobación.
- **Individuo vs. Organización:**
  - *Individuo*: te inscribes con tu nombre; ese nombre aparece como firmante ("TU NOMBRE").
    Rápido de activar.
  - *Organización*: aparece el nombre legal de la empresa; **requiere un número D-U-N-S** y prueba
    de que la entidad legal existe y de que tienes autoridad para representarla. Tarda más.
- **Requisitos de cuenta:** Apple ID con 2FA obligatorio; aceptar los acuerdos vigentes.

---

## 5. Ventajas y desventajas

**Ventajas**
- Instalación sin fricción ni advertencias en macOS; percepción de confianza/profesionalismo.
- El escaneo de notarización da una señal (débil, automatizada) de que el binario no es malware
  conocido.
- Requisito de facto para Homebrew casks serios y para cualquier distribución directa amplia.

**Desventajas / costos**
- **99 USD/año** recurrentes mientras quieras firmar builds nuevos.
- Complejidad operativa: certificados, claves, rotaciones, runners macOS en CI (minutos 10×).
- Dependencia de Apple: si Apple revoca tu cuenta/cert, tu canal de distribución firmado se corta.
- No sustituye a auditar tu propio código: la notarización es automática, no una revisión de
  seguridad real.

**Alternativas si no quieres pagar (por ahora)**
- **Distribuir sin firmar** y documentar el "Abrir de todos modos" o `xattr -dr com.apple.quarantine ./andromeda`. UX peor, pero cero costo.
- **Homebrew formula desde fuente** (`brew install --build-from-source`): el usuario compila local;
  no cruza Gatekeeper porque no es un binario descargado. Requiere Go instalado en la máquina del
  usuario.
- **Firma ad-hoc** (`codesign -s -`): elimina algunos avisos pero **no** pasa notarización; no es
  distribución de confianza.

---

## 6. Términos legales y jurisdicción

- El **Apple Developer Program License Agreement (ADPLA)** es el contrato que aceptas. Está
  redactado en inglés y (en sus versiones estándar) se **rige por las leyes del Estado de
  California, EE. UU.**, con resolución de disputas allí — independientemente de tu país. Léelo: es
  extenso y cambia con el tiempo.
- **Tu país de desarrollo/residencia** impone además: obligaciones fiscales (declarar ingresos si
  monetizas), posible registro como autónomo/empresa, IVA/impuestos sobre ventas si cobras, y reglas
  de protección de datos si el software procesa datos personales. **Esto no lo asumo aquí**; depende
  de dónde estés constituido y conviene confirmarlo con un contador/abogado local.
- **Exportación/criptografía:** Andromeda hace TLS y maneja credenciales. La mayoría de librerías
  estándar caen en excepciones, pero la distribución de software con criptografía tiene reglas de
  export (EE. UU. EAR y equivalentes locales). Para un CLI open-source con cripto estándar el riesgo
  práctico es bajo, pero es un punto a verificar si comercializas a gran escala.

---

## 7. Propiedad intelectual y monetización

- **Firmar/notarizar NO transfiere propiedad intelectual.** Tu código sigue siendo tuyo bajo su
  licencia (Andromeda es **Apache-2.0**, ver ADR-002). Apple no adquiere derechos sobre tu código;
  tú le concedes derechos **limitados** (por el ADPLA) para operar el servicio de notarización y,
  si publicas en App Store, para distribuir tu app.
- **¿Es monetizable?** Sí, y notarizar **no** te obliga a un modelo:
  - **Distribución directa** (web/GitHub/Homebrew) notarizada: te quedas el **100%**, pero tú
    gestionas cobros, impuestos y soporte. Apple no cobra comisión por binarios notarizados fuera de
    la App Store.
  - **Mac App Store**: Apple retiene **15–30%** de comisión (15% en el *Small Business Program* si
    facturas menos de 1 M USD/año). Requiere sandbox de App Store, que para un CLI/TUI que ejecuta
    procesos y toca el sistema es **muy restrictivo** — probablemente no encaje con Andromeda.
  - **Modelo open-core / servicios**: al ser Apache-2.0, puedes cobrar por soporte, hosting,
    features enterprise, o binarios firmados "listos para usar" mientras el core sigue abierto. La
    licencia Apache-2.0 **permite** uso comercial (tuyo y de terceros); no impide monetizar.
- **Marca:** "andromeda" y el gato mascota son tu identidad de marca (ADR-026). Considera registrar
  la marca si comercializas; eso es independiente de Apple y depende de tu país/EUIPO/USPTO.

---

## 8. Checklist para activar

- [ ] Inscripción Apple Developer Program (99 USD/año) — individuo u organización.
- [ ] Certificado *Developer ID Application* exportado a `.p12` + contraseña.
- [ ] API Key de App Store Connect (`.p8`, Key ID, Issuer ID).
- [ ] Secretos en CI: `MACOS_SIGN_P12`, `MACOS_SIGN_PASSWORD`, `MACOS_NOTARY_KEY`,
      `MACOS_NOTARY_KEY_ID`, `MACOS_NOTARY_ISSUER_ID`.
- [ ] Prueba local de firma + `notarytool submit --wait` OK.
- [ ] `spctl -a -vvv` → *accepted / Notarized* en una Mac limpia.
- [ ] Decidir cadencia (firmar solo en tags de versión para ahorrar minutos macOS en CI).

Una vez con esto, el `release.yml` + `.goreleaser.yaml` existentes producen artefactos firmados y
notarizados sin cambios de código: el bloque ya está escrito y se enciende con las credenciales.
