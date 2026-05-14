cmake_minimum_required(VERSION 3.20)
project(poc01_capture CXX)

set(CMAKE_CXX_STANDARD 17)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# Windows-only PoC
if(NOT WIN32)
  message(FATAL_ERROR "PoC 01 requires Windows (DXGI + NVENC)")
endif()

# NVENC SDK path — set NVENC_SDK_PATH or install CUDA Toolkit
set(NVENC_SDK_PATH "" CACHE PATH "Path to NVIDIA NVENC SDK headers")

add_executable(poc01_capture src/main.cpp)

target_include_directories(poc01_capture PRIVATE
  ${NVENC_SDK_PATH}/Interface
)

target_link_libraries(poc01_capture PRIVATE
  d3d11
  dxgi
  dxguid
)

target_compile_definitions(poc01_capture PRIVATE
  NOMINMAX
  WIN32_LEAN_AND_MEAN
)
