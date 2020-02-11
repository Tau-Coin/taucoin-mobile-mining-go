Pod::Spec.new do |spec|
  spec.name         = 'Gtau'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/Tau-Coin/taucoin-mobile-mining-go'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Tau Client'
  spec.source       = { :git => 'https://github.com/Tau-Coin/taucoin-mobile-mining-go.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Gtau.framework'

	spec.prepare_command = <<-CMD
    curl https://gtaustore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Gtau.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
