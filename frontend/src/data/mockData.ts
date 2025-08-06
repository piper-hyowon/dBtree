import {Database, DatabaseDetail} from '../types/dashboard.types';

export const mockDatabases: Database[] = [
    {
        id: "daebef54-7352-41e2-ba16-5b03b979c22a",
        name: "mydb",
        type: "mongodb",
        size: "small",
        mode: "standalone",
        status: "provisioning",
        resources: {
            cpu: 2,
            memory: 2048,
            disk: 20
        },
        cost: {
            creationCost: 20,
            hourlyLemons: 2,
            dailyLemons: 48,
            monthlyLemons: 1440
        },
        backupEnabled: false,
        config: {
            version: "7.0",
            wiredTigerCache: 1
        },
        createdAt: "2025-08-06T04:11:13.337219+09:00",
        updatedAt: "2025-08-06T04:11:13.337219+09:00",
        createdFromPreset: "mongodb-standalone-small"
    },
    {
        id: "abc123-7352-41e2-ba16-5b03b979c22a",
        name: "production-db",
        type: "mongodb",
        size: "medium",
        mode: "replica",
        status: "running",
        resources: {
            cpu: 4,
            memory: 4096,
            disk: 100
        },
        cost: {
            creationCost: 50,
            hourlyLemons: 5,
            dailyLemons: 120,
            monthlyLemons: 3600
        },
        backupEnabled: true,
        config: {
            version: "15.0"
        },
        createdAt: "2025-07-06T04:11:13.337219+09:00",
        updatedAt: "2025-08-06T04:11:13.337219+09:00",
        createdFromPreset: "postgresql-replica-medium"
    },
    {
        id: "def456-7352-41e2-ba16-5b03b979c22a",
        name: "mongomongo",
        type: "mongodb",
        size: "small",
        mode: "standalone",
        status: "stopped",
        resources: {
            cpu: 1,
            memory: 1024,
            disk: 10
        },
        cost: {
            creationCost: 10,
            hourlyLemons: 1,
            dailyLemons: 24,
            monthlyLemons: 720
        },
        backupEnabled: false,
        config: {
            version: "7.0"
        },
        createdAt: "2025-06-06T04:11:13.337219+09:00",
        updatedAt: "2025-08-06T04:11:13.337219+09:00",
        createdFromPreset: "redis-standalone-small"
    }
];

export const mockDatabaseDetails: { [key: string]: DatabaseDetail } = {
    "daebef54-7352-41e2-ba16-5b03b979c22a": {
        ...mockDatabases[0],
        externalHost: "db-cluster-01.dbtree.io",
        externalPort: 27017,
        externalUriTemplate: "mongodb://{USERNAME}:{PASSWORD}@db-cluster-01.dbtree.io:27017/mydb?authSource=admin"
    },
    "abc123-7352-41e2-ba16-5b03b979c22a": {
        ...mockDatabases[1],
        externalHost: "pg-cluster-02.dbtree.io",
        externalPort: 5432,
        externalUriTemplate: "postgresql://{USERNAME}:{PASSWORD}@pg-cluster-02.dbtree.io:5432/production-db"
    },
    "def456-7352-41e2-ba16-5b03b979c22a": {
        ...mockDatabases[2],
        externalHost: "redis-cluster-03.dbtree.io",
        externalPort: 6379,
        externalUriTemplate: "redis://:{PASSWORD}@redis-cluster-03.dbtree.io:6379/0"
    }
};

export const lemonCredits = 120;