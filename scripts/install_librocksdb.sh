#!/bin/bash

ROCKSDB_VERSION=${1:-"9.3.1"}

# Check if RocksDB is already installed
if [[ $(uname) == "Darwin" ]]; then
	# On macOS, check for RocksDB in /opt/homebrew/lib
	if [[ -f "/opt/homebrew/lib/librocksdb.dylib" ]]; then
		read -r -p "RocksDB is already installed in /opt/homebrew/lib. Do you want to reinstall it? (yes/no): " choice
		case "$choice" in
		y | yes | Yes | YES)
			echo "Reinstalling RocksDB..."
			rm -rf /opt/homebrew/lib/librocksdb* /opt/homebrew/include/rocksdb
			;;
		n | no | No | NO)
			echo "Skipping RocksDB installation."
			exit 0
			;;
		*)
			echo "Invalid choice. Please enter 'yes' or 'no'."
			exit 1
			;;
		esac
	else
		echo "RocksDB is not installed. Proceeding with installation..."
	fi
else
	# On Linux, check for RocksDB in /usr/lib
	if [[ $(find /usr/lib -name "librocksdb.so.${ROCKSDB_VERSION}" -print -quit) ]]; then
		read -r -p "RocksDB version ${ROCKSDB_VERSION} is already installed. Do you want to reinstall it? (yes/no): " choice
		case "$choice" in
		y | yes | Yes | YES)
			echo "Reinstalling RocksDB..."
			rm -rf /usr/lib/librocksdb*
			;;
		n | no | No | NO)
			echo "Skipping RocksDB installation."
			exit 0
			;;
		*)
			echo "Invalid choice. Please enter 'yes' or 'no'."
			exit 1
			;;
		esac
	else
		echo "RocksDB is not installed. Proceeding with installation..."
	fi
fi

# Clean up temporary files
rm -rf /tmp/rocksdb

# Check the OS type and perform different actions
if [[ $(uname) == "Linux" ]]; then
	# Linux installation process
	if [[ -f /etc/os-release ]]; then
		source /etc/os-release
		if [[ "$ID" == "ubuntu" ]]; then
			# Install dependencies for Ubuntu
			echo "Installing RocksDB dependencies..."
			apt-get install libgflags-dev libsnappy-dev zlib1g-dev libbz2-dev liblz4-dev libzstd-dev build-essential clang
		elif [[ "$ID" == "alpine" ]]; then
			# Install dependencies for Alpine
			echo "Installing RocksDB dependencies..."
			echo "@testing http://nl.alpinelinux.org/alpine/edge/testing" >>/etc/apk/repositories
			apk add --update --no-cache cmake bash perl g++
			apk add --update --no-cache zlib zlib-dev bzip2 bzip2-dev snappy snappy-dev lz4 lz4-dev zstd@testing zstd-dev@testing libtbb-dev@testing libtbb@testing
			# Install latest gflags
			cd /tmp &&
				git clone https://github.com/gflags/gflags.git &&
				cd gflags &&
				mkdir build &&
				cd build &&
				cmake -DBUILD_SHARED_LIBS=1 -DGFLAGS_INSTALL_SHARED_LIBS=1 .. &&
				make install &&
				rm -rf /tmp/gflags
		else
			echo "Linux distribution not supported"
			exit 1
		fi
		# Build and install RocksDB from source for any Linux distribution
		cd /tmp &&
			git clone -b v"${ROCKSDB_VERSION}" --single-branch https://github.com/facebook/rocksdb.git &&
			cd rocksdb &&
			PORTABLE=1 WITH_JNI=0 WITH_BENCHMARK_TOOLS=0 WITH_TESTS=1 WITH_TOOLS=0 WITH_CORE_TOOLS=1 WITH_BZ2=1 WITH_LZ4=1 WITH_SNAPPY=1 WITH_ZLIB=1 WITH_ZSTD=1 WITH_GFLAGS=0 USE_RTTI=1 \
				make shared_lib &&
			cp librocksdb.so* /usr/lib/ &&
			cp -r include/* /usr/include/ &&
			rm -rf /tmp/rocksdb
	else
		echo "Cannot determine Linux distribution."
		exit 1
	fi

elif [[ $(uname) == "Darwin" ]]; then
	# macOS installation process
	xcode-select --install # Ensure Xcode command line tools are installed
	# brew install rocksdb does not support specifying version, so build from source
	echo "brew install rocksdb does not support version selection, compiling from source v${ROCKSDB_VERSION}..."
	brew install gcc snappy lz4 zstd zlib bzip2 || true # Install gcc if not present
	cd /tmp && \
		git clone -b v"${ROCKSDB_VERSION}" --single-branch https://github.com/facebook/rocksdb.git && \
		cd rocksdb && \
		PORTABLE=1 WITH_JNI=0 WITH_BENCHMARK_TOOLS=0 WITH_TESTS=1 WITH_TOOLS=0 WITH_CORE_TOOLS=1 WITH_BZ2=1 WITH_LZ4=1 WITH_SNAPPY=1 WITH_ZLIB=1 WITH_ZSTD=1 WITH_GFLAGS=0 USE_RTTI=1 \
			make shared_lib && \
		cp librocksdb.dylib* /opt/homebrew/lib/ && \
		cp -r include/* /opt/homebrew/include/ && \
        ln -sf librocksdb.dylib /opt/homebrew/lib/librocksdb.9.3.dylib
		rm -rf /tmp/rocksdb
else
	echo "Unsupported OS."
	exit 1
fi