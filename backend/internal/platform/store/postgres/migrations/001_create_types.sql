-- User
CREATE TYPE session_status AS ENUM ('pending', 'verified');

-- Lemon
CREATE TYPE lemon_action AS ENUM (
    'welcome_bonus',
    'harvest',
    'instance_create',
    'instance_maintain'
    );
CREATE TYPE transaction_status AS ENUM ('successful', 'failed');

-- Quiz
CREATE TYPE quiz_difficulty AS ENUM ('easy', 'normal', 'hard');
CREATE TYPE quiz_category AS ENUM ('basics', 'sql', 'design', 'performance', 'cloud');
CREATE TYPE quiz_status AS ENUM ('started', 'done', 'timeout');
CREATE TYPE harvest_status AS ENUM ('none', 'in_progress', 'success', 'failure', 'timeout');

-- DB Instance
CREATE TYPE db_type AS ENUM ('mongodb', 'redis');
CREATE TYPE db_size AS ENUM ('small', 'medium', 'large');
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