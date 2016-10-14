# Author: Antoine Mercadal
# See LICENSE file for full LICENSE
# Copyright 2016 Aporeto.

import os
import shutil

from monolithe.generators.lib import TemplateFileWriter


class APIVersionWriter(TemplateFileWriter):
    """ This class is reponsible to write files for a particular api version. """

    def __init__(self, monolithe_config, api_info):
        """
        """
        super(APIVersionWriter, self).__init__(package="monobahamut")

        output = monolithe_config.get_option("output", "transformer")
        self.output_directory = "%s/bahamut/%s" % (output, api_info["version"])

        self.base_package = monolithe_config.get_option("base_package", "bahamut")
        self.models_package_package = monolithe_config.get_option("models_package_package", "bahamut")
        self.handlers_package_name = monolithe_config.get_option("handlers_package_name", "bahamut")
        self.routes_package_name = monolithe_config.get_option("routes_package_name", "bahamut")
        self.models_package_name = monolithe_config.get_option("models_package_name", "bahamut")

        handlers_folder = "%s/handlers" % self.output_directory
        if os.path.exists(handlers_folder):
            shutil.rmtree(handlers_folder)
        os.makedirs(handlers_folder)

        routes_folder = "%s/routes" % self.output_directory
        if os.path.exists(routes_folder):
            shutil.rmtree(routes_folder)
        os.makedirs(routes_folder)

        code_header_path = "%s/bahamut/__code_header" % output
        if os.path.exists(code_header_path):
            with open(code_header_path, "r") as f:
                self.header_content = f.read()

    def perform(self, specifications):
        """
        """
        self._write_handler_config()
        for rest_name, specification in specifications.iteritems():
            self._write_handler(specification=specification)

        self._write_routes(specifications=specifications)
        self._format()

    def _write_handler_config(self):
        """
        """
        filename = 'handlers/handler_config.go'
        self.write(destination=self.output_directory, filename=filename, template_name="handler_config.go.tpl",
                   handlers_package_name=self.handlers_package_name)

    def _write_handler(self, specification):
        """
        """
        filename = 'handlers/%s_handler.go' % specification.rest_name

        self.write(destination=self.output_directory, filename=filename, template_name="handler.go.tpl",
                   specification=specification,
                   base_package=self.base_package,
                   handlers_package_name=self.handlers_package_name,
                   routes_package_name=self.routes_package_name,
                   models_package_package=self.models_package_package,
                   models_package_name=self.models_package_name)

    def _write_routes(self, specifications):
        """
        """
        filename = 'routes/routes.go'

        self.write(destination=self.output_directory, filename=filename, template_name="routes.go.tpl",
                   specifications=specifications,
                   base_package=self.base_package,
                   handlers_package_name=self.handlers_package_name,
                   routes_package_name=self.routes_package_name,
                   models_package_package=self.models_package_package,
                   models_package_name=self.models_package_name)

    def _format(self):
        """
        """
        os.system("gofmt -w '%s' >/dev/null 2>&1" % self.output_directory)
