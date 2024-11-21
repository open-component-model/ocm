{{- /* Go template for Homebrew Formula */ -}}
# typed: false
# frozen_string_literal: true

class {{ classname }} < Formula
  desc "The OCM CLI makes it easy to create component versions and embed them in build processes."
  homepage "https://ocm.software/"
  version "{{ .Version }}"

  on_macos do
    on_intel do
      url "{{ .ReleaseURL }}/v{{ .Version }}/ocm-{{ .Version }}-darwin-amd64.tar.gz"
      sha256 "{{ .darwin_amd64_sha256 }}"

      def install
        bin.install "ocm"
      end
    end
    on_arm do
      url "{{ .ReleaseURL }}/v{{ .Version }}/ocm-{{ .Version }}-darwin-arm64.tar.gz"
      sha256 "{{ .darwin_arm64_sha256 }}"

      def install
        bin.install "ocm"
      end
    end
  end

  on_linux do
    on_intel do
      if Hardware::CPU.is_64_bit?
        url "{{ .ReleaseURL }}/v{{ .Version }}/ocm-{{ .Version }}-linux-amd64.tar.gz"
        sha256 "{{ .linux_amd64_sha256 }}"

        def install
          bin.install "ocm"
        end
      end
    end
    on_arm do
      if Hardware::CPU.is_64_bit?
        url "{{ .ReleaseURL }}/v{{ .Version }}/ocm-{{ .Version }}-linux-arm64.tar.gz"
        sha256 "{{ .linux_arm64_sha256 }}"

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
