# Instalación e invocación de `andromeda`

Cómo se logra que `andromeda` se ejecute en cualquier terminal y cómo funciona el instalador
rápido tipo `curl … | bash`.

---

## 1. ¿Cómo funciona "escribir `andromeda` en cualquier terminal"?

No hay magia: cuando escribes `andromeda`, el shell busca un ejecutable llamado `andromeda` en los
directorios listados en tu variable **`PATH`**. Que funcione "en cualquier terminal" solo significa
que **el binario está en un directorio que está en el PATH**. Hay tres formas de conseguirlo:

| Método | Qué hace | Ideal para |
|---|---|---|
| **Homebrew** (`brew install`) | Instala el binario en `/opt/homebrew/bin` (Apple Silicon) o `/usr/local/bin`, que ya están en el PATH | La mayoría de usuarios macOS/Linux |
| **Instalador `curl \| bash`** | Descarga el binario del release y lo copia a `/usr/local/bin` o `~/.local/bin` | Instalación rápida sin gestor de paquetes |
| **`go install`** | Compila desde fuente a `$(go env GOBIN)` (o `~/go/bin`) | Desarrolladores con Go |
| **Manual** | Descargas el `.tar.gz`, extraes `andromeda`, `chmod +x`, lo mueves a un dir del PATH | Control total |

Hoy, para desarrollo, ya lo tienes: `make build` produce `./bin/andromeda`. Para invocarlo como
`andromeda` a secas, ese `./bin` tiene que estar en el PATH, o copiar el binario a `/usr/local/bin`:

```bash
make build
sudo cp ./bin/andromeda /usr/local/bin/andromeda   # ya invocable como: andromeda
# o sin sudo:
mkdir -p ~/.local/bin && cp ./bin/andromeda ~/.local/bin/ && echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.zshrc
```

---

## 2. El instalador rápido (`curl -fsSL … | bash`)

Ya existe: **`scripts/install.sh`**. Detecta tu SO/arquitectura, resuelve la última release, descarga
el `.tar.gz` correcto de **GitHub Releases**, extrae `andromeda` y lo instala en el PATH.

```bash
# Forma pública (cuando el repo/artefactos sean accesibles sin auth):
curl -fsSL https://raw.githubusercontent.com/datamaia/andromeda/main/scripts/install.sh | bash

# Con opciones:
ANDROMEDA_VERSION=v0.1.0 ANDROMEDA_INSTALL_DIR="$HOME/.local/bin" \
  curl -fsSL https://raw.githubusercontent.com/datamaia/andromeda/main/scripts/install.sh | bash
```

El ejemplo de Hermes (`curl -fsSL https://hermes-agent.nousresearch.com/install.sh | bash`) es
exactamente esto, con una diferencia: ellos sirven el `install.sh` desde **su propio dominio** en
vez de `raw.githubusercontent.com`. Es cosmético — el script hace lo mismo.

---

## 3. ¿Qué se requiere para hostearlo? (esto es lo importante)

Dos artefactos distintos que hay que servir:

**A) El binario (los releases).** Ya está resuelto: `.goreleaser.yaml` + `.github/workflows/release.yml`
construyen los `.tar.gz` por plataforma, checksums, SBOM y firmas cosign, y los **publican como un
GitHub Release** cuando haces push de un tag `vX.Y.Z`. No necesitas un servidor propio: **GitHub
Releases es el hosting**. El instalador descarga de ahí.

```bash
git tag v0.1.0 && git push origin v0.1.0     # dispara el release; goreleaser sube los artefactos
```

**B) El `install.sh`.** Necesita una URL estable. Opciones:
1. **`raw.githubusercontent.com/.../install.sh`** — cero infraestructura. Es la más simple.
2. **Un dominio propio** (como Hermes): un redirect/hosting estático (GitHub Pages, Cloudflare,
   Vercel) que sirva el mismo archivo en `https://tudominio/install.sh`. Solo estética/branding.

> **No** necesitas "alojar el artefacto en el repo" (no metas binarios en git). El flujo correcto es
> repo con el **código** → tag → CI construye → **GitHub Releases** guarda los binarios.

---

## 4. El bloqueante real: repositorio privado

Aquí está el detalle que hay que decidir. El repo hoy es **privado** en una cuenta **gratuita**, y
eso rompe el `curl | bash` público en dos puntos:

- **`raw.githubusercontent.com/...install.sh`** de un repo privado **exige autenticación** → un
  `curl` anónimo da 404. No puedes compartir un one-liner que "solo funcione".
- **Descargar el binario del Release** de un repo privado también exige un token.

Es decir: la instalación tipo Hermes (un comando público que cualquiera pega) **requiere que los
releases sean accesibles sin auth**. Opciones:

| Opción | Instalador público funciona | Costo | Notas |
|---|---|---|---|
| **Hacer el repo público** | ✅ sí | Gratis | Expone el código (es open-source Apache-2.0 de todas formas). Habilita además branch protection gratis. |
| **Mantener privado + servir binarios aparte** | ✅ sí | Hosting propio | Subes los `.tar.gz` a un bucket/CDN público (R2, S3, Releases de un repo público espejo) y el `install.sh` apunta allí. Más piezas. |
| **Mantener privado, instalación con token** | ⚠️ solo con `GITHUB_TOKEN` | Gratis | El instalador ya soporta `GITHUB_TOKEN=...`; sirve para ti/tu equipo, no para el público. |

El instalador (`scripts/install.sh`) ya contempla las tres: usa `GITHUB_TOKEN` si está, y si no,
va anónimo (que funciona en cuanto los releases sean públicos).

---

## 5. Recomendación

Como Andromeda es **open-source (Apache-2.0)** por diseño, lo natural es **hacer el repositorio
público** cuando llegues a la primera release. Con eso, en un solo movimiento:

- el `curl | bash` público funciona sin tokens,
- **branch protection y rulesets pasan a estar disponibles gratis** (hoy bloqueados por ser privado
  en plan free),
- Homebrew tap y `go install` funcionan para cualquiera.

Si aún no quieres exponerlo, el instalador con `GITHUB_TOKEN` te cubre para uso propio/equipo hasta
que decidas publicar.

---

## 6. Resumen de lo que ya está y lo que falta

- ✅ `scripts/install.sh` — instalador con detección de plataforma, `GITHUB_TOKEN` opcional, PATH.
- ✅ `.goreleaser.yaml` + `release.yml` — construyen y publican releases (el hosting es GitHub Releases).
- ⏳ **Publicar la primera release** (`git tag v0.1.0 && git push --tags`) para que haya algo que descargar.
- ⏳ **Decidir público vs. privado** (§4) — condiciona el `curl | bash` público y branch protection.
- ⏳ (Opcional) dominio propio para el `install.sh` (estético).
