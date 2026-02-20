# This formula is auto-updated by GoReleaser on each tagged release.
# Manual edits will be overwritten.
#
# For local development, build and install directly:
#   make build && make install

class Spark < Formula
  desc "Workspace CLI for multi-repo development"
  homepage "https://github.com/BrianBFarias/homebrew-spark"
  license "MIT"
  version "0.1.0"

  on_macos do
    if Hardware::CPU.arm?
      url "https://github.com/BrianBFarias/homebrew-spark/releases/download/v0.1.0/spark_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/BrianBFarias/homebrew-spark/releases/download/v0.1.0/spark_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  on_linux do
    if Hardware::CPU.arm?
      url "https://github.com/BrianBFarias/homebrew-spark/releases/download/v0.1.0/spark_linux_arm64.tar.gz"
      sha256 "PLACEHOLDER"
    else
      url "https://github.com/BrianBFarias/homebrew-spark/releases/download/v0.1.0/spark_linux_amd64.tar.gz"
      sha256 "PLACEHOLDER"
    end
  end

  def install
    bin.install "spark"
  end

  test do
    system "#{bin}/spark", "version"
  end
end
