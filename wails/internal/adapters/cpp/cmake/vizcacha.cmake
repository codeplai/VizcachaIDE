# Written by VizcachaIDE and loaded with CMAKE_PROJECT_TOP_LEVEL_INCLUDES: on Windows it adds the
# UTF-8 console helper to every executable of the project, so accents work in the console.
function(vizcacha_add_console)
  get_property(targets DIRECTORY "${CMAKE_SOURCE_DIR}" PROPERTY BUILDSYSTEM_TARGETS)
  foreach(target IN LISTS targets)
    get_target_property(type ${target} TYPE)
    if(type STREQUAL "EXECUTABLE")
      target_sources(${target} PRIVATE "vizcacha_console_utf8.cpp")
    endif()
  endforeach()
endfunction()
if(WIN32)
  cmake_language(DEFER DIRECTORY "${CMAKE_SOURCE_DIR}" CALL vizcacha_add_console)
endif()
