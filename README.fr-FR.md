<div align="center">

<img src="nutshell-icon.svg" width="80" height="80" alt="icône nutshell" />

# nutshell

**Un standard ouvert pour empaqueter le contexte de tâches que les agents IA peuvent comprendre.**

Fonctionne avec n'importe quel agent : Claude Code · Copilot · Cursor · OpenClaw · Agents personnalisés

[Spécification](spec/nutshell-spec-v0.2.0.md) · [Exemples](examples/) · [Recherche](docs/harness-engineering-research.md) · [Site web](https://chatchat.space/nutshell/)

[English](README.md) | [简体中文](README.zh-CN.md) | [繁體中文](README.zh-HANT.md) | [Español](README.es-ES.md) | **[Français](README.fr-FR.md)**

</div>

---

## Le Problème

Les agents de programmation IA sont puissants, mais ils posent toujours les mêmes questions :

```
Agent : "Quel framework ? Quelle base de données ? Où est le schéma ?
         Comment s'authentifier ? Quels sont les critères d'acceptation ?
         Est-ce que je peux accéder à l'environnement de staging ?"
Humain : *envoie 47 messages en 3 jours, perdant le contexte à chaque fois*
```

Chaque fois que vous démarrez une nouvelle session, vous ré-expliquez le même contexte. Les identifiants sont partagés via Slack. Les exigences n'existent que dans votre tête. Il n'y a aucune trace de ce qui a été fait ni pourquoi.

## La Solution

**Nutshell** empaquète tout ce dont un agent IA a besoin dans un seul paquet :

```
$ nutshell init
$ nutshell check

  🐚 Vérification de Complétude Nutshell

  ✓ task.title: "Construire une REST API pour la Gestion des Utilisateurs"
  ✓ task.summary: fourni
  ✓ context/requirements.md: existe (2.1 Ko)
  ✗ context/architecture.md: référencé mais manquant
  ✗ credentials: pas de coffre — l'agent n'aura pas accès à la BD
  ⚠ acceptance: pas de scripts de test — l'agent ne peut pas s'auto-vérifier

  Statut : INCOMPLET — 2 éléments nécessitent attention avant que l'agent puisse commencer
```

Nutshell vous dit **à vous** ce qui manque. Comblez les lacunes, empaquetez et transmettez à n'importe quel agent :

```
$ nutshell pack -o task.nut       # L'humain empaquète la tâche
$ nutshell inspect task.nut       # L'agent voit tout ce dont il a besoin
# ... l'agent exécute ...
$ nutshell pack -o delivery.nut   # L'agent livre les résultats
```

---

## Pourquoi Nutshell ?

| Sans Nutshell | Avec Nutshell |
|--------------|--------------|
| Contexte dispersé entre Slack, docs, emails | Un seul paquet `.nut` avec tout |
| L'agent pose 20 questions avant de commencer | L'agent lit le manifeste, commence immédiatement |
| Identifiants partagés de manière non sécurisée | Coffre chiffré avec jetons limités en portée et en temps |
| Aucun historique des demandes ou livraisons | Les paquets demande + livraison forment une piste d'audit complète |
| Nouvelle session = tout ré-expliquer | Le paquet persiste entre les sessions |
| Aucun moyen de vérifier l'achèvement | Critères d'acceptation lisibles par machine |

### Conception Autonome

Nutshell fonctionne **sans aucune plateforme externe**. Un seul développeur avec Claude Code en bénéficie immédiatement :

1. **Définir** — `nutshell init` crée un répertoire de tâches structuré
2. **Vérifier** — `nutshell check` vous dit ce qui manque (identifiants ? docs d'architecture ? critères d'acceptation ?)
3. **Empaqueter** — `nutshell pack` compresse en un paquet `.nut`
4. **Exécuter** — Transmettez le paquet à n'importe quel agent IA
5. **Archiver** — Les paquets de livraison documentent ce qui a été construit et pourquoi

### Extensions de Plateforme (Optionnelles)

Vous voulez publier des tâches sur un marché ? Nutshell supporte des extensions optionnelles :

```jsonc
{
  "extensions": {
    "clawnet": {                    // Réseau P2P d'agents
      "peer_id": "12D3KooW...",
      "reward": {"amount": 50, "currency": "energy"}
    },
    "linear": {"issue_id": "ENG-1234"},
    "github-actions": {"workflow": "agent-task.yml"}
  }
}
```

Les extensions ne cassent jamais le format de base. Les outils ignorent ce qu'ils ne comprennent pas.

---

## 🐚 Le Nom

> **龍蝦吃貝殼** — *Les homards mangent des coquillages.*

[ClawNet](https://github.com/ChatChatTech/ClawNet) (🦞) est un réseau décentralisé d'agents IA. Les agents sont des homards. Ils ont besoin de nourriture — et la nourriture vient dans des coquilles. **Nutshell** (🐚) est la coquille — compacte, nutritive, prête à être ouverte.

Mais vous n'avez pas besoin d'être un homard. N'importe quel agent peut manger un nutshell.

---

## Démarrage Rapide

### Installer

```bash
# Installation en une ligne (détecte automatiquement l'OS et l'architecture)
curl -fsSL https://chatchat.space/nutshell/install.sh | sh

# Ou via Go
go install github.com/ChatChatTech/nutshell/cmd/nutshell@latest

# Ou compiler depuis les sources
git clone https://github.com/ChatChatTech/nutshell.git
cd nutshell && make build
```

### Créer une Tâche

```bash
# Initialiser
nutshell init --dir my-task
cd my-task

# Éditer le manifeste
vim nutshell.json

# Vérifier ce qui manque
nutshell check

# Empaqueter quand c'est prêt
nutshell pack -o my-task.nut
```

### Inspecter un Paquet

```
$ nutshell inspect my-task.nut

    🐚  n u t s h e l l  🦞
    Empaquetage de Tâches pour Agents IA

  Bundle: my-task.nut
  Version: 0.2.0
  Type: request
  ID: nut-7f3a1b2c-...

  📋 Tâche : Construire une REST API pour la Gestion des Utilisateurs
  Priorité : high | Effort : 8h

  🏷️  Étiquettes : golang, postgresql, jwt, rest-api
  Domaines : backend, authentication

  👤 Éditeur : Alice Chen (via claude-code)

  🔑 Identifiants : 2 avec portée
    • staging-db (postgresql) — read-write
    • api-token (bearer_token) — invoke

  📦 Fichiers : 5 fichiers, 8 200 octets

  ⚙️  Indices Harness :
    Type d'agent : execution
    Stratégie : incremental
    Budget de contexte : 0.35
```

### Valider

```bash
nutshell validate my-task.nut      # vérifier le paquet empaqueté
nutshell validate ./my-task        # vérifier le répertoire
```

### Édition Rapide

```bash
nutshell set task.title "Build REST API"
nutshell set task.priority high
nutshell set tags.skills_required "go,rest,api"
```

### Comparer des Paquets

```bash
nutshell diff request.nut delivery.nut          # différence lisible par l'humain
nutshell diff request.nut delivery.nut --json   # différence lisible par machine
```

### JSON Schema

```bash
nutshell schema                            # afficher sur stdout
nutshell schema -o nutshell.schema.json    # écrire dans un fichier
```

Ajouter à `nutshell.json` pour l'auto-complétion IDE :
```jsonc
{
  "$schema": "./schema/nutshell.schema.json",
  ...
}
```

### Commandes Avancées

```bash
# Compression contextuelle — analyse les types de fichiers et applique la compression optimale
nutshell compress --dir ./my-task -o task.nut --level best

# Division de paquets multi-agents — divise une tâche en sous-tâches parallèles
nutshell split --dir ./my-task -n 3
nutshell merge part-0/ part-1/ part-2/ -o merged/

# Rotation des identifiants — auditer et mettre à jour l'expiration des identifiants
nutshell rotate --dir ./my-task                              # auditer tous
nutshell rotate staging-db --expires 2026-01-01T00:00:00Z    # faire tourner un seul

# Visionneuse web — visionneuse HTTP locale pour l'inspection de .nut
nutshell serve ./my-task --port 8080
nutshell serve task.nut
```

---

## Structure du Paquet

```
task.nut                        🐚 La coquille
├── nutshell.json               📋 Manifeste (toujours chargé en premier)
├── context/                    📖 Exigences, architecture, références
├── files/                      📦 Fichiers sources et ressources
├── apis/                       🔌 Spécifications d'API appelables
├── credentials/                🔑 Coffre d'identifiants chiffré
├── tests/                      ✅ Critères d'acceptation et scripts de test
└── delivery/                   🦪 Artefacts de finalisation (paquets de livraison)
```

Seul `nutshell.json` est obligatoire. Ajoutez des répertoires selon les besoins.

## Manifeste (`nutshell.json`)

```jsonc
{
  "nutshell_version": "0.2.0",
  "bundle_type": "request",
  "id": "nut-a1b2c3d4-...",
  "task": {
    "title": "Construire une REST API pour la gestion des utilisateurs",
    "summary": "Endpoints CRUD avec authentification JWT et PostgreSQL.",
    "priority": "high",
    "estimated_effort": "8h"
  },
  "tags": {
    "skills_required": ["golang", "postgresql", "jwt"],
    "domains": ["backend"],
    "custom": {"framework": "gin"}
  },
  "publisher": {
    "name": "Alice Chen",
    "tool": "claude-code"
  },
  "context": {
    "requirements": "context/requirements.md",
    "architecture": "context/architecture.md"
  },
  "credentials": {
    "vault": "credentials/vault.enc.json",
    "encryption": "age",
    "scopes": [
      {"name": "staging-db", "type": "postgresql", "access_level": "read-write", "expires_at": "2026-03-21T10:00:00Z"}
    ]
  },
  "acceptance": {
    "checklist": [
      "Tous les endpoints CRUD retournent les bons codes de statut",
      "L'authentification JWT fonctionne pour les routes protégées"
    ],
    "auto_verifiable": true
  },
  "harness": {
    "agent_type_hint": "execution",
    "context_budget_hint": 0.35,
    "execution_strategy": "incremental",
    "constraints": ["Ne pas modifier les fichiers en dehors de files/src/"]
  },
  "completeness": {
    "status": "ready"
  }
}
```

Seuls `nutshell_version`, `bundle_type`, `id` et `task.title` sont obligatoires. Tout le reste améliore l'efficacité de l'agent.

---

## La Commande Check (Gestion Inversée)

La fonctionnalité la plus puissante : **Nutshell gère l'humain**.

```bash
$ nutshell check

  🐚 Vérification de Complétude Nutshell

  ✓ task.title: "Build REST API"
  ✓ context/requirements.md: existe (2.1 Ko)
  ✗ context/architecture.md: référencé mais manquant
  ✗ credentials: pas de coffre — l'agent n'aura pas accès à la BD
  ⚠ acceptance: pas de critères — l'agent ne peut pas s'auto-vérifier
  ⚠ harness: pas de contraintes

  Statut : INCOMPLET — comblez 2 éléments avant que l'agent puisse commencer
```

Au lieu que l'agent demande « de quoi d'autre ai-je besoin ? », le **paquet dit à l'humain** quoi fournir. Cela inverse la dynamique habituelle et garantit que les agents reçoivent un contexte complet dès le départ.

---

## Alignement Harness Engineering

Nutshell est fondé sur [Harness Engineering](docs/harness-engineering-research.md) — la discipline émergente de construction d'infrastructure autour des agents IA :

| Principe | Implémentation Nutshell |
|----------|------------------------|
| **Architecture de Contexte** | Chargement par niveaux — manifeste d'abord, détails à la demande |
| **Spécialisation des Agents** | `harness.agent_type_hint` guide quel rôle d'agent convient |
| **Mémoire Persistante** | Les paquets de livraison préservent les logs d'exécution, décisions, points de contrôle |
| **Exécution Structurée** | Séparation demande/livraison avec critères d'acceptation lisibles par machine |
| **Règle des 40%** | `context_budget_hint` empêche le débordement de la fenêtre de contexte |
| **Mécanisation des Contraintes** | Les contraintes Harness sont lisibles par machine et applicables |

---

## Sécurité des Identifiants

| Principe | Implémentation |
|----------|---------------|
| **Limité en portée** | Chaque identifiant restreint à des tables, endpoints, actions spécifiques |
| **Limité en temps** | Chaque identifiant a un `expires_at` |
| **Chiffré** | Par défaut : [chiffrement age](https://age-encryption.org/). Supporte aussi SOPS, Vault |
| **Limité en débit** | Limites de débit par identifiant |
| **Auditable** | Les paquets de livraison enregistrent quels identifiants ont été utilisés |

---

## Intégration ClawNet

Nutshell s'intègre nativement avec [ClawNet](https://github.com/ChatChatTech/ClawNet) — un réseau décentralisé de communication entre agents. Les deux projets sont **entièrement indépendants** (zéro dépendance à la compilation), mais utilisés ensemble ils fournissent un flux de travail fluide publier → réclamer → livrer sur un réseau P2P.

### Prérequis

- Un daemon ClawNet en cours d'exécution (`clawnet start`) sur `localhost:3998`
- Nutshell CLI (ce projet)

### Flux de Travail

```bash
# 1. L'auteur crée un paquet de tâches et le publie sur le réseau
nutshell init --dir my-task
#    ... remplir nutshell.json, ajouter les fichiers de contexte ...
nutshell publish --dir my-task

# 2. Un autre agent parcourt et réclame la tâche
nutshell claim <task-id> -o workspace/

# 3. L'agent termine le travail et livre
nutshell deliver --dir workspace/
```

### Ce qui se passe en coulisses

| Étape | Nutshell | ClawNet |
|-------|----------|---------|
| `publish` | Empaquète le paquet `.nut`, mappe manifeste → champs de tâche | Crée la tâche dans le Task Bazaar, stocke le paquet, diffuse aux pairs |
| `claim` | Télécharge le paquet `.nut` (ou crée depuis les métadonnées) | Retourne les détails de la tâche + blob du paquet |
| `deliver` | Empaquète le paquet de livraison, soumet le résultat | Met à jour le statut de la tâche à `submitted`, stocke le paquet de livraison |

### Schéma d'Extension

Les tâches publiées stockent les métadonnées ClawNet dans `extensions.clawnet` :

```json
{
  "extensions": {
    "clawnet": {
      "peer_id": "12D3KooW...",
      "task_id": "a1b2c3d4-...",
      "reward": 10.0
    }
  }
}
```

### Adresse ClawNet Personnalisée

```bash
nutshell publish --clawnet http://192.168.1.5:3998 --dir my-task
nutshell claim --clawnet http://remote:3998 <task-id>
```

---

## Exemples

| Exemple | Description | Type |
|---------|------------|------|
| [01-api-task](examples/01-api-task/) | Tâche de développement REST API | Demande |
| [02-data-analysis](examples/02-data-analysis/) | Analyse de données avec S3 | Demande |
| [03-delivery](examples/03-delivery/) | Livraison terminée | Livraison |

---

## Spécification

Spécification complète : [spec/nutshell-spec-v0.2.0.md](spec/nutshell-spec-v0.2.0.md)

Sections clés :
- §2 Structure du Paquet
- §3 Schéma du Manifeste
- §4 Vérification de Complétude
- §5 Schéma de Livraison
- §6 Système d'Étiquettes
- §7 Coffre d'Identifiants
- §8 Format de Spécification d'API
- §9 Critères d'Acceptation
- §10 Extensions (ClawNet, GitHub Actions, etc.)
- §11 Type MIME
- §12 Gestion de Version

---

## Feuille de Route

- [x] v0.2.0 — Spécification autonome-d'abord
- [x] Go CLI (`init`, `pack`, `unpack`, `inspect`, `validate`, `check`, `set`, `diff`, `schema`)
- [x] Paquets d'exemple (demande + livraison)
- [x] JSON Schema pour l'auto-complétion IDE
- [x] `nutshell set` — Édition rapide des champs du manifeste via notation à points
- [x] `nutshell diff` — Comparer paquets de demande vs livraison
- [x] Checksums SHA-256 au niveau fichier
- [x] Types de paquets étendus (template, checkpoint, partial)
- [x] Agent SDK — `nutshell.Open()` API Go pour accès programmatique aux paquets
- [x] Intégration native ClawNet (`publish`, `claim`, `deliver` via P2P Task Bazaar)
- [x] Compression contextuelle (Nutcracker Phase 2)
- [x] Extension VS Code pour l'édition de paquets
- [x] Division de paquets multi-agents (sous-tâches parallèles)
- [x] Protocole de rotation des identifiants
- [x] Visionneuse web pour l'inspection de `.nut`

---

## Contribuer

Nutshell est un standard ouvert. Les contributions sont les bienvenues :

1. **Améliorations de la spécification** — Ouvrez un issue ou PR contre `spec/`
2. **Exemples** — Ajoutez des exemples réels de paquets à `examples/`
3. **Outillage** — Construisez des intégrations pour votre framework d'agents
4. **Extensions** — Définissez de nouveaux schémas d'extension pour votre plateforme

---

## Licence

MIT

---

<div align="center">

**🐚 Empaqueter. Ouvrir. Expédier.**

*Un standard ouvert par [ChatChatTech](https://github.com/ChatChatTech)*

</div>
