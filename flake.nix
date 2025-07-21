{
  description = "A Nix-based development and production environment for PeakStreak";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; overlays = []; };
        
        go-migrate-pg = pkgs.go-migrate.overrideAttrs(oldAttrs: {
          tags = ["postgres"];
        });
      in
      {
        devShells.default = pkgs.mkShell {
          name = "peakstreak-dev-env";
          
          packages = [
            pkgs.go_1_24
            pkgs.gopls
            pkgs.delve
            pkgs.bun
            pkgs.postgresql_17
            go-migrate-pg
            pkgs.gnumake
            pkgs.git
          ];

          shellHook = ''
            echo "--- PeakStreak Development Environment ---"

            trap 'echo "Stopping PostgreSQL..."; pg_ctl -D ./db-data -l postgres.log stop 2>/dev/null' EXIT

            export PGDATA=$PWD/db-data
            export DB_URL="postgres://user:password@localhost:5432/peakstreak?sslmode=disable"
            export JWT_SECRET="a-secure-secret-for-development"
            export SERVER_PORT="8080"
            
            if [ ! -d "$PGDATA" ]; then
              echo "Initializing PostgreSQL database in $PGDATA..."
              initdb -D $PGDATA --no-locale --encoding=UTF8 -U user
            fi

            if ! pg_ctl -D $PGDATA status > /dev/null; then
              echo "Starting PostgreSQL..."
              pg_ctl -D $PGDATA -l postgres.log start
              
              until pg_isready -q -h localhost -p 5432 -U user; do
                echo "Waiting for PostgreSQL to start..."
                sleep 1
              done
            else
              echo "PostgreSQL is already running."
            fi

            psql -U user -lqt postgres | cut -d \| -f 1 | grep -q3 peakstreak || createdb -U user -O user peakstreak

            echo ""
            echo "✅ PostgreSQL is running."
            echo "   (It will be stopped automatically when you exit this shell)"
            echo "   DB URL: $DB_URL"
            echo "✅ Environment is ready."
            echo ""
            echo "Next steps:"
            echo "1. Run 'make migrate-up' to apply database migrations."
            echo "2. In one terminal: 'make run' (for the backend)"
            echo "3. In a second terminal: 'make dev-frontend' (for frontend hot-reloading)"
            echo ""
          '';
        };
      }
    );
}
