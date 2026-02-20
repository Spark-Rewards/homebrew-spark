# spark-cli

Multi-repo workspace CLI for Spark Rewards development.

**New to spark-cli?** See **[docs/GETTING_STARTED.md](docs/GETTING_STARTED.md)** for a guide (purpose, install, and how to use it).

## Install

```bash
brew install spark-rewards/spark-cli/spark-cli 
```

## Usage

```bash
spark-cli create workspace ~/SparkRewards
cd ~/SparkRewards
spark-cli use AppModel
spark-cli use AppAPI
spark-cli sync
cd AppAPI && spark-cli run build
```

Run `spark-cli --help` for all commands.
