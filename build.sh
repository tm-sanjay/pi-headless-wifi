#!/bin/bash
Version=1.0

build_package() {
    echo "Building package for $1"
    ARCH=$1

    # Create a new directory for the package
    mkdir -p myapp-deb/DEBIAN
    mkdir -p myapp-deb/usr/local/bin

    # create build directory
    mkdir -p build

    # Build the Go application
    if [ $ARCH = "amd64" ]; then
        env GOOS=linux GOARCH=arm GOARM=7 go build -o myapp
    else
        GOOS=linux GOARCH=arm64 go build -o myapp
    fi
    
    # Copy the binary to the package directory
    cp myapp myapp-deb/usr/local/bin/

    # Create the control file
    cat << EOF > myapp-deb/DEBIAN/control
Package: myapp
Version: $Version
Section: base
Priority: optional
Architecture: $ARCH
Maintainer: Your Name <you@example.com>
Description: A short description of your app
EOF

    # Set the permissions on the binary
    chmod 755 myapp-deb/usr/local/bin/myapp

    # Build the package
    dpkg-deb --build myapp-deb

    # rename the file with version
    mv myapp-deb.deb build/myapp-$Version-$ARCH.deb

    # Clean up
    rm -rf myapp-deb

    echo "Package built for $1"
}

build_package "amd64"
# build_package "arm64"
