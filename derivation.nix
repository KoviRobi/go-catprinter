{
  buildGoModule,
  src,
  version,
  cups,
}:
buildGoModule {
  pname = "tooltracker";
  inherit src version;

  vendorHash = "sha256-3McvjMMnfZ2VLhlnW+ENmYbHOCzwjtrRKxO2Tq8pPcg=";

  # To speed up build -- tests are more for development than packaging
  doCheck = false;

  postInstall = ''
    mkdir -p $out/lib/cups/filter/
    mv $out/bin/rastertocatprinter $out/lib/cups/filter/

    mkdir -p $out/share/cups/model/
    ${cups}/bin/ppdc -d $out/share/cups/model/ cups/catprinter.drv
  '';

  meta.mainProgram = "catprinter";
}
