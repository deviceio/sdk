package net.deviceio.agent;

import android.app.IntentService;
import android.content.Intent;

import agent.Agent;

/**
 * Responsible for setting up the agent transport connection as a intent service. This SHOULD NOT be
 * run under the UI thread, as the bindings and call to the agent golang code will block the UI thread
 * causing the app to go unresponsive.
 */
public class TransportService extends IntentService {
    /**
     * The agent id to use when connecting to the hub
     */
    public static String agentId;

    /**
     * The hub ip or hostname to connect to
     */
    public static String hubHost;

    /**
     * The hub port to use when connecting
     */
    public static int hubPort = 8975;

    /**
     * Constructor
     */
    public TransportService() {
        super("TransportService");
    }

    /**
     * Called when the intent is handled
     * @param workIntent
     */
    @Override
    protected void onHandleIntent(Intent workIntent) {
        Agent.start(agentId, hubHost, hubPort, true);
    }
}
