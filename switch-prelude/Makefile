TARGET := dying-light-prelude
BUILD := build
SOURCES := source
INCLUDES :=

ARCH := -march=armv8-a+crc+crypto -mtune=cortex-a57 -mtp=soft -fPIE
CFLAGS := -g -Wall -Wextra -O2 -ffunction-sections -fdata-sections $(ARCH) $(INCLUDES)
CFLAGS += -D__SWITCH__
CXXFLAGS := $(CFLAGS) -fno-rtti -fno-exceptions
ASFLAGS := -g $(ARCH)
LDFLAGS := -specs=$(DEVKITPRO)/libnx/switch.specs -g -Wl,-Map,$(notdir $*.map) -Wl,--gc-sections
LIBS := -lnx

include $(DEVKITPRO)/libnx/switch_rules
