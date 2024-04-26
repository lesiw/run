if run.argv[1] == "" then
  run.argv[1] = "build"
end
if run.argv[1] == "build" then
  run.env["RUNCTR"] = "./etc/Dockerfile.dev"
end
