{
  buildGoModule,
  src,
  version,
  cups,
}:
buildGoModule {
  pname = "go-catprinter";
  inherit src version;

  vendorHash = "sha256-bFCfnDKiUKjcznFbL2XzATOPmz/I98FZ2dlJYCjus4g=";

  # To speed up build -- tests are more for development than packaging
  doCheck = false;

  postInstall = ''
    mkdir -p $out/lib/cups/backend/
    mv $out/bin/cupsbackend $out/lib/cups/backend/catprinter

    mkdir -p $out/share/cups/model/
    ${cups}/bin/ppdc -d $out/share/cups/model/ cups/catprinter.drv
  '';

  meta.mainProgram = "catprinter";
}
