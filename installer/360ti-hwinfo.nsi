; 360ti HWiNFO - NSIS installer
; Compile with NSIS >= 3.0 (https://nsis.sourceforge.io)
;   makensis 360ti-hwinfo.nsi
; Expected file layout (relative to this script):
;   ..\dist\360ti-hwinfo.exe
;   ..\dist\config.json
;   ..\dist\logo360ti.png

Unicode true

!include "MUI2.nsh"

; -------------------------------------------------------- metadata
Name "360ti HWiNFO"
OutFile "360ti-hwinfo-setup.exe"
InstallDir "$PROGRAMFILES64\360ti\Analisys"
InstallDirRegKey HKLM "Software\360ti\Analisys" "InstallDir"
RequestExecutionLevel admin
SetCompressor /SOLID lzma

; -------------------------------------------------------- interface
!define MUI_ABORTWARNING
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!define MUI_FINISHPAGE_RUN "$INSTDIR\360ti-hwinfo.exe"
!insertmacro MUI_PAGE_FINISH
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_LANGUAGE "PortugueseBR"

; -------------------------------------------------------- sections
Section "Aplicação" MainSec
    SetOutPath "$INSTDIR"

    File "..\dist\360ti-hwinfo.exe"
    File "..\dist\config.json"
    File "..\dist\logo360ti.png"

    ; uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    ; shortcuts
    CreateDirectory "$SMPROGRAMS\360ti"
    CreateShortcut "$SMPROGRAMS\360ti\360ti HWiNFO.lnk" "$INSTDIR\360ti-hwinfo.exe"
    CreateShortcut "$SMPROGRAMS\360ti\Desinstalar 360ti HWiNFO.lnk" "$INSTDIR\Uninstall.exe"
    CreateShortcut "$DESKTOP\360ti HWiNFO.lnk" "$INSTDIR\360ti-hwinfo.exe"

    ; registry (add/remove programs)
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "DisplayName" "360ti HWiNFO"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "DisplayVersion" "1.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "Publisher" "360ti"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "DisplayIcon" "$INSTDIR\360ti-hwinfo.exe,0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "InstallLocation" "$INSTDIR"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "QuietUninstallString" '"$INSTDIR\Uninstall.exe" /S'
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "NoModify" 1
    WriteRegDWORD HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO" "NoRepair" 1
    WriteRegStr HKLM "Software\360ti\Analisys" "InstallDir" "$INSTDIR"
SectionEnd

; -------------------------------------------------------- uninstall
Section "Uninstall"
    Delete "$INSTDIR\360ti-hwinfo.exe"
    Delete "$INSTDIR\config.json"
    Delete "$INSTDIR\logo360ti.png"
    Delete "$INSTDIR\Uninstall.exe"
    RMDir /r "$INSTDIR\reports"
    RMDir "$INSTDIR"

    Delete "$SMPROGRAMS\360ti\360ti HWiNFO.lnk"
    Delete "$SMPROGRAMS\360ti\Desinstalar 360ti HWiNFO.lnk"
    RMDir "$SMPROGRAMS\360ti"
    Delete "$DESKTOP\360ti HWiNFO.lnk"

    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\360ti HWiNFO"
    DeleteRegKey HKLM "Software\360ti\Analisys"
SectionEnd