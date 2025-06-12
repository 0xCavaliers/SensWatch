# Build script for sens_match project

# Create necessary directories if they don't exist
if (-not (Test-Path "bin")) {
    New-Item -ItemType Directory -Path "bin"
}
if (-not (Test-Path "lib")) {
    New-Item -ItemType Directory -Path "lib"
}

# Copy Python script
Copy-Item "../address_name.py" -Destination "lib/"

# Build Go program
Write-Host "Building Go program..."
Set-Location src
go build -o ../bin/sens_match.exe
Set-Location ..

# Build MD5 calculator
Write-Host "Building MD5 calculator..."
Set-Location ../read_path
gcc md5_calculator.c -o ../sens_match/bin/md5_calculator.exe -I"vcpkg安装路径/installed/x64-mingw-dynamic/include" -L"vcpkg安装路径/installed/x64-mingw-dynamic/lib" -lssl -lcrypto
Copy-Item "libcrypto-3-x64.dll" -Destination "../sens_match/bin/"
Copy-Item "libssl-3-x64.dll" -Destination "../sens_match/bin/"
Set-Location ../sens_match

Write-Host "Build completed!" 