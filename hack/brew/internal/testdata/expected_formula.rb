# typed: false
# frozen_string_literal: true

class OcmAT100 < Formula
  desc "The OCM CLI makes it easy to create component versions and embed them in build processes."
  homepage "https://ocm.software/"
  version "1.0.0"

  on_macos do
    on_intel do
      url "$$TEST_SERVER$$/v1.0.0/ocm-1.0.0-darwin-amd64.tar.gz"
      sha256 "dummy-digest"

      def install
        bin.install "ocm"
      end
    end
    on_arm do
      url "$$TEST_SERVER$$/v1.0.0/ocm-1.0.0-darwin-arm64.tar.gz"
      sha256 "dummy-digest"

      def install
        bin.install "ocm"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "$$TEST_SERVER$$/v1.0.0/ocm-1.0.0-linux-amd64.tar.gz"
        sha256 "dummy-digest"

        def install
          bin.install "ocm"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "$$TEST_SERVER$$/v1.0.0/ocm-1.0.0-linux-arm64.tar.gz"
        sha256 "dummy-digest"

        def install
          bin.install "ocm"
        end
      end
    end
  end

  test do
    system "#{bin}/ocm --version"
  end
end
