@echo off
echo Building protoc-gen-mongo plugin...

REM 创建bin目录
if not exist "bin" mkdir bin

REM 构建插件
go build -o bin/protoc-gen-mongo.exe ./cmd/protoc-gen-mongo
if %ERRORLEVEL% neq 0 (
    echo Build failed!
    exit /b 1
)

echo Plugin built successfully!

@REM REM 检查protoc是否可用
@REM protoc --version >nul 2>&1
@REM if %ERRORLEVEL% neq 0 (
@REM     echo Warning: protoc not found in PATH. Please install Protocol Buffers compiler.
@REM     echo Download from: https://github.com/protocolbuffers/protobuf/releases
@REM )

@REM echo.
@REM echo To use the plugin, add the bin directory to your PATH, or use:
@REM echo protoc --plugin=protoc-gen-mongo=./bin/protoc-gen-mongo.exe --mongo_out=./output your_file.proto

pause 