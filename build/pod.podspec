Pod::Spec.new do |spec|
  spec.name         = 'Glemo'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/LemoFoundationLtd/lemochain-go'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Lemochain Client'
  spec.source       = { :git => 'https://github.com/LemoFoundationLtd/lemochain-go.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Glemo.framework'

	spec.prepare_command = <<-CMD
    curl https://glemostore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Glemo.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
