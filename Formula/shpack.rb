class Shpack < Formula
  desc "Shell Script Bundler - package multiple scripts into a single executable"
  homepage "https://github.com/luongnguyen1805/shpack"
  url "https://github.com/luongnguyen1805/shpack/releases/download/1.0.0/shpack-1.0.0.tar.gz"
  sha256 "d8b69ee6cf9434a8fe7005b1be2c98aada99b5090e233bb20548b0b3173406ea"
  license "MIT"
  version "1.0.0"

  def install
    bin.install "shpack"
  end

  test do
    system "#{bin}/shpack", "version"
  end
end
