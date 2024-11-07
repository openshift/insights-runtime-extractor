cargo vendor > ./.cargo/config.toml

# Vendoring with cargo will pull all libraries to build on Windows.
# As this is not needed, we remove manually all dependencies to Windows

find ./vendor/windows*/src/ ! -name 'lib.rs' -type f -exec rm -f {} +
find ./vendor/winapi*/src/ ! -name 'lib.rs' -type f -exec rm -f {} +
rm -fr ./vendor/windows*/lib/*.a
rm -fr ./vendor/winapi*/lib/*.a
rm -fr ./vendor/winapi*/lib/*.lib
rm -fr ./vendor/windows*/lib/*.lib

