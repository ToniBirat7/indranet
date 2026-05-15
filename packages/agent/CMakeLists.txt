cmake_minimum_required(VERSION 3.20)
project(indranet-agent VERSION 0.1.0 LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)

# vcpkg integration
if(DEFINED ENV{VCPKG_ROOT})
  set(CMAKE_TOOLCHAIN_FILE "$ENV{VCPKG_ROOT}/scripts/buildsystems/vcpkg.cmake")
endif()

# Find dependencies
find_package(nlohmann_json CONFIG REQUIRED)
find_package(websocketpp CONFIG REQUIRED)

# Optional: NVENC SDK (set NVENC_SDK_PATH)
if(DEFINED NVENC_SDK_PATH)
  include_directories(${NVENC_SDK_PATH}/Interface)
  add_compile_definitions(HAVE_NVENC)
endif()

# Optional: AMF SDK (set AMF_SDK_PATH)
if(DEFINED AMF_SDK_PATH)
  include_directories(${AMF_SDK_PATH}/amf/public/include)
  add_compile_definitions(HAVE_AMF)
endif()

set(AGENT_SOURCES
  src/main.cpp
  src/capture/dxgi_capture.cpp
  src/encode/nvenc_encoder.cpp
  src/stream/webrtc_streamer.cpp
  src/input/vigem_input.cpp
  src/sandbox/windows_sandbox.cpp
  src/billing/billing_client.cpp
  src/ipc/agent_ipc.cpp
)

add_executable(indranet-agent ${AGENT_SOURCES})

target_include_directories(indranet-agent PRIVATE
  ${CMAKE_CURRENT_SOURCE_DIR}/src
)

target_link_libraries(indranet-agent PRIVATE
  nlohmann_json::nlohmann_json
  websocketpp::websocketpp
)

if(WIN32)
  target_link_libraries(indranet-agent PRIVATE
    d3d11
    dxgi
    dxguid
  )
  # ViGEm client SDK — expects VIGEM_SDK_PATH
  if(DEFINED VIGEM_SDK_PATH)
    target_include_directories(indranet-agent PRIVATE ${VIGEM_SDK_PATH}/include)
    target_link_directories(indranet-agent PRIVATE ${VIGEM_SDK_PATH}/lib/driver/MSVC/x64)
    target_link_libraries(indranet-agent PRIVATE ViGEmClient)
  endif()
endif()

install(TARGETS indranet-agent DESTINATION bin)
