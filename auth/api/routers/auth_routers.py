from api.auth.login import login_bpp
from api.auth.registers import register_bpp
from api.auth.refresh import refresh_bpp
from api.auth.logout import logout_bpp

def register_all_blueprints(app):
    app.register_blueprint(login_bpp)
    app.register_blueprint(register_bpp)
    app.register_blueprint(refresh_bpp)
    app.register_blueprint(logout_bpp)