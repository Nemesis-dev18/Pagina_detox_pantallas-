require "sinatra/base"

class DetoxWeb < Sinatra::Base
  set :public_folder, File.expand_path("..", __dir__)
  set :views, File.expand_path("views", __dir__)
  enable :sessions

  helpers do
    def page_title(title = nil)
      title || "DETOX_WEB"
    end
  end

  get "/" do
    @page = :index
    erb :index
  end

  get "/inicio_sesion" do
    @page = :inicio_sesion
    erb :inicio_sesion
  end

  post "/inicio_sesion" do
    redirect "/inicio_sesion"
  end

  get "/registro" do
    @page = :registro
    erb :registro
  end

  post "/registro" do
    redirect "/registro"
  end

  get "/recuperar_contraseña" do
    @page = :recuperar
    erb :recuperar
  end

  post "/recuperar_contraseña" do
    redirect "/recuperar_contraseña"
  end

  get "/nosotros" do
    @page = :nosotros
    erb :nosotros
  end

  get "/mapa" do
    @page = :mapa
    erb :mapa
  end

  get "/comunidad" do
    @page = :comunidad
    erb :comunidad
  end

  get "/test" do
    @page = :test
    erb :test
  end
end

DetoxWeb.run! if __FILE__ == $PROGRAM_NAME
