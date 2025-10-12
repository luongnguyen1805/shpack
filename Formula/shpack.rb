class Shpack < Formula
  desc "Shell Script Bundler - package multiple scripts into a single executable"
  homepage "https://github.com/luongnguyen1805/shpack"
  license "None"
  version "1.0.0"

  on_macos do
    if Hardware::CPU.intel?
      url "https://luongnguyen1805.github.io/shpack/v1.0.0/shpack-darwin-amd64.tar.gz", :using => :nounzip
      sha256 "TO_BE_UPDATED_darwin-amd64"
    else
      url "https://luongnguyen1805.github.io/shpack/v1.0.0/shpack-darwin-arm64.tar.gz", :using => :nounzip
      sha256 "d8b69ee6cf9434a8fe7005b1be2c98aada99b5090e233bb20548b0b3173406ea"
    end
  end

  on_linux do
    if Hardware::CPU.intel?
      url "https://luongnguyen1805.github.io/shpack/v1.0.0/shpack-linux-amd64.tar.gz", :using => :nounzip
      sha256 "TO_BE_UPDATED_linux-amd64"
    else
      url "https://luongnguyen1805.github.io/shpack/v1.0.0/shpack-linux-arm64.tar.gz", :using => :nounzip
      sha256 "TO_BE_UPDATED_linux-arm64"
    end
  end

  def install
    # Fix: Extract tarball and install binary as 'shpack'
    system "tar", "-xzf", Dir["*.tar.gz"].first
    bin.install Dir["shpack-*"].first => "shpack"
  end

  test do
    system "#{bin}/shpack", "--version"
  end
end