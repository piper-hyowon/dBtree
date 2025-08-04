-- User
DO
$$
    BEGIN
        CREATE TYPE session_status AS ENUM ('pending', 'verified');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

-- Lemon
DO
$$
    BEGIN
        CREATE TYPE lemon_action AS ENUM (
            'welcome_bonus',
            'harvest',
            'instance_create',
            'instance_maintain'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE transaction_status AS ENUM ('successful', 'failed');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

-- Quiz
DO
$$
    BEGIN
        CREATE TYPE quiz_difficulty AS ENUM ('easy', 'normal', 'hard');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE quiz_category AS ENUM ('basics', 'sql', 'design', 'performance', 'cloud');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE quiz_status AS ENUM ('started', 'done', 'timeout');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE harvest_status AS ENUM ('none', 'in_progress', 'success', 'failure', 'timeout');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

-- DB Instance
DO
$$
    BEGIN
        CREATE TYPE db_type AS ENUM ('mongodb', 'redis');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE db_size AS ENUM ('small', 'medium', 'large');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE db_mode AS ENUM ('standalone', 'replica_set', 'sharded', 'basic', 'sentinel', 'cluster');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE db_status AS ENUM (
            'provisioning',
            'running',
            'paused',
            'stopped',
            'error',
            'deleting',
            'maintenance',
            'backing_up',
            'restoring',
            'upgrading'
            );
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE billing_status AS ENUM ('pending', 'processed', 'failed', 'cancelled');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE backup_type AS ENUM ('manual', 'scheduled');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;

DO
$$
    BEGIN
        CREATE TYPE backup_status AS ENUM ('pending', 'running', 'completed', 'failed');
    EXCEPTION
        WHEN duplicate_object THEN null;
    END
$$;